package projects

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/sergi/go-diff/diffmatchpatch"

	"github.com/jchorl/collabtest/constants"
	"github.com/jchorl/collabtest/models"
)

func run(c echo.Context) error {
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project run")
		return errors.New("Unable to get DB from context")
	}

	hash := c.Param("hash")

	// Verify that the proj exists
	if db.First(&models.Project{Hash: hash}).RecordNotFound() {
		return constants.UNRECOGNIZED_HASH
	}

	// Get docker client for communicating with docker
	dockerClient, ok := c.Get(constants.CTX_DOCKER_CLIENT).(*client.Client)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get docker client from context in project test run")
		return errors.New("Unable to get docker client from context")
	}

	file, err := c.FormFile("file")
	if err != nil {
		logrus.WithError(err).Error("Unable to get file from upload req")
		return err
	}

	// Get upload program to test
	testProgramReader, err := file.Open()
	if err != nil {
		logrus.WithError(err).Error("Could not open uploaded file to write to disk")
		return err
	}
	defer testProgramReader.Close()

	// Get container configuration for type of program submitted
	containerConfig, err := selectConfig(file.Filename)
	if err != nil {
		return err
	}

	// Make sure the cmd pipes in the test input
	containerConfig.Cmd[len(containerConfig.Cmd)-1] = containerConfig.Cmd[len(containerConfig.Cmd)-1] + " < testIn"

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Resources: container.Resources{
			CPUShares: constants.BUILD_CPU_SHARE,
			Memory:    constants.BUILD_MEMORY,
		},
	}

	// Get project test files
	testFiles, err := ioutil.ReadDir(path.Join("projects", hash))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"dir":   path.Join("projects", hash),
		}).Error("unable to ls dir to get test files to run")
	}

	// Run program through each test and record expected and actual result diffs
	diffs := []string{}
	for _, testFile := range testFiles {
		if !strings.HasSuffix(testFile.Name(), ".in") {
			continue
		}

		logrus.WithField("name", testFile.Name()).Debug("processing test file")

		// Create a container to run test
		createResponse, err := dockerClient.ContainerCreate(context.Background(), &containerConfig, hostConfig, nil, "")
		if err != nil {
			logrus.WithError(err).Error("Could not create container")
			return err
		}

		// Copy test program into container
		err = copyToContainer(context.Background(), dockerClient, testProgramReader, createResponse.ID, path.Join("/build", file.Filename))
		if err != nil {
			logrus.WithError(err).Error("Could not copy executable to container")
			return err
		}

		// Reset test program reader for next run
		_, err = testProgramReader.Seek(0, io.SeekStart)
		if err != nil {
			logrus.WithError(err).Error("Could not seek to start of program reader")
			return err
		}

		// Get test input file
		testFileReader, err := os.Open(path.Join("projects", hash, testFile.Name()))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"hash":  hash,
				"file":  path.Join("projects", hash, testFile.Name()),
			}).Error("unable to open test file to run tests")
			continue
		}
		defer testFileReader.Close()

		// copy test input file into container
		err = copyToContainer(context.Background(), dockerClient, testFileReader, createResponse.ID, path.Join("/build", "testIn"))
		if err != nil {
			logrus.WithError(err).Error("Could not copy test file to container")
			return err
		}

		// Run container
		if err := dockerClient.ContainerStart(context.Background(), createResponse.ID, types.ContainerStartOptions{}); err != nil {
			logrus.WithError(err).Error("Error starting container")
			return err
		}

		// Wait until container finishes execution
		dockerClient.ContainerWait(context.Background(), createResponse.ID)

		// Get logs from container which include program output
		logsOptions := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
		}
		logsReadCloser, err := dockerClient.ContainerLogs(context.Background(), createResponse.ID, logsOptions)
		logsBuffer := new(bytes.Buffer)
		if _, err := logsBuffer.ReadFrom(logsReadCloser); err != nil {
			logrus.WithError(err).Error("Cannot read container logs")
		}
		defer logsReadCloser.Close()

		// Docker prepends 8 bytes of info: see HEADER at https://docs.docker.com/v1.6/reference/api/docker_remote_api_v1.14/
		logsBuffer.Next(8)

		// Save a instance of the project being run
		runInstance := models.Run{
			Project: models.Project{Hash: hash},
			Stdout:  logsBuffer.String(),
			Stderr:  "",
		}

		db.Create(&runInstance)

		expectedOutFilename := testFile.Name()[0:len(testFile.Name())-len(filepath.Ext(testFile.Name()))] + ".out" // disgusting way of removing extension. TODO improve.
		// Open expected test file output
		testFileOutputReader, err := os.Open(path.Join("projects", hash, expectedOutFilename))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"hash":  hash,
				"file":  path.Join("projects", hash, expectedOutFilename),
			}).Error("unable to open test output file while running tests")
			continue
		}
		defer testFileOutputReader.Close()

		// Diff expected and actual output
		d, err := diff(testFileOutputReader, logsBuffer)
		if err != nil {
			logrus.WithError(err).Error("unable to diff expected output with actual output")
			continue
		}

		// Add diff of actual and expected output
		diffs = append(diffs, d)
	}

	return c.JSON(http.StatusOK, diffs)
}

// Helper for copying file from memory directly into running docker container without first saving to filesystem
func copyToContainer(ctx context.Context, dockerClient *client.Client, file io.ReadCloser, dstContainer, dstPath string) (err error) {
	// Create a pipe with one end passed to docker function, and the other end passed to buffer
	pipeReader, pipeWriter := io.Pipe()
	buf := bufio.NewWriterSize(pipeWriter, 32*1024)

	go func() {
		// Create a tar writer such that all writes are written to buf
		tarWriter := tar.NewWriter(buf)
		// Create a buffer that flushes to the tar writer
		tarBuf := bufio.NewWriterSize(tarWriter, 32*1024)

		defer func() {
			// Flushing buf writes to pipeWriter
			buf.Flush()
			if err := tarWriter.Close(); err != nil {
				logrus.WithError(err).Error("Can't close tar writer")
			}
			if err := pipeWriter.Close(); err != nil {
				logrus.WithError(err).Error("Can't close pipe writer")
			}
		}()

		// Write contents of file into tarWriter
		if err := addTarFile(filepath.Base(dstPath), file, tarWriter, tarBuf); err != nil {
			// If pipe is broken, stop writing tar stream to it
			if err == io.ErrClosedPipe {
				logrus.WithError(err).Error("Error adding file to tar due to closed pipe")
				return
			} else {
				logrus.WithError(err).Error("Error adding file to tar")
				return
			}
		}
	}()

	return dockerClient.CopyToContainer(ctx, dstContainer, filepath.Dir(dstPath), pipeReader, types.CopyToContainerOptions{})
}

// Help function to tar file so docker client's CopyToContainer method can be used
func addTarFile(filename string, file io.ReadCloser, tarWriter *tar.Writer, tarBuf *bufio.Writer) error {
	buffer := new(bytes.Buffer)
	size, err := buffer.ReadFrom(file)
	if err != nil {
		logrus.WithError(err).Error("Error reading into buffer")
		return err
	}

	// Tar header
	hdr := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: size,
	}

	// Write tar header
	if err := tarWriter.WriteHeader(hdr); err != nil {
		logrus.WithError(err).Error("Error writing header to tar file")
		return err
	}

	// Copy file buffer into tarBuff
	_, err = io.Copy(tarBuf, buffer)
	if err != nil {
		logrus.WithError(err).Error("Error copying from buffer to tar buffer")
		return err
	}

	// Flushing tarBuf writes to tarWriter
	err = tarBuf.Flush()
	if err != nil {
		logrus.WithError(err).Error("Error flushing tar buffer")
		return err
	}

	return nil
}

func diffSample(c echo.Context) error {
	in1, err := os.Open("projects/incorrect/output/test.out")
	if err != nil {
		logrus.WithError(err).Error("Error opening in1 for reading.")
		return err
	}
	in2, err := os.Open("projects/incorrect/expected/test.out")
	if err != nil {
		logrus.WithError(err).Error("Error opening in2 for reading.")
		return err
	}

	d, err := diff(in1, in2)
	if err != nil {
		logrus.WithError(err).Error("Error computing file diff.")
		return err
	}
	logrus.Debug(d)
	return c.HTML(http.StatusOK, d)
}

// Diffs 2 io.Readers
func diff(in1, in2 io.Reader) (string, error) {
	// Read io.Readers into strings
	buf := new(bytes.Buffer)
	buf.ReadFrom(in1)
	str1 := buf.String()

	buf.Reset()
	buf.ReadFrom(in2)
	str2 := buf.String()

	logrus.WithFields(logrus.Fields{
		"str1": str1,
		"str2": str2,
	}).Debug("diffing")

	// Diff strings
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(str1, str2, false)
	return dmp.DiffPrettyHtml(diffs), nil
}

// Select the correct docker config depending on the type of program being run
func selectConfig(srcpath string) (container.Config, error) {
	// Get file extension
	var extension = filepath.Ext(srcpath)
	logrus.Debug(extension)

	timeout := constants.BUILD_TIMEOUT

	config := container.Config{
		WorkingDir:      "/build",
		NetworkDisabled: true,
		StopTimeout:     &timeout,
	}

	// Get container config based on file extension
	defaults, ok := constants.FILETYPE_CONFIGS[extension]
	if !ok {
		logrus.WithField("extension", extension).Error("Attempting to run file with unknown extension")
		return container.Config{}, fmt.Errorf("Unsupported file extension: %s", extension)
	}

	config.Image = defaults.Image()
	config.Cmd = defaults.Command(srcpath)

	return config, nil
}
