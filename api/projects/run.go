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

	// verify that the proj exists
	if db.First(&models.Project{}, hash).RecordNotFound() {
		return constants.UNRECOGNIZED_HASH
	}

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

	testProgramReader, err := file.Open()
	if err != nil {
		logrus.WithError(err).Error("Could not open uploaded file to write to disk")
		return err
	}
	defer testProgramReader.Close()

	containerConfig, err := selectConfig(file.Filename)
	if err != nil {
		return err
	}

	// make sure the cmd pipes in the test input
	containerConfig.Cmd = append(containerConfig.Cmd, "<", "testIn")

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Resources: container.Resources{
			CPUShares: constants.BUILD_CPU_SHARE,
			Memory:    constants.BUILD_MEMORY,
		},
	}

	// get project test files
	testFiles, err := ioutil.ReadDir(path.Join("projects", hash))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"dir":   path.Join("projects", hash),
		}).Error("unable to ls dir to get test files to run")
	}

	diffs := []string{}
	for _, testFile := range testFiles {
		if !strings.HasSuffix(testFile.Name(), ".in") {
			continue
		}

		logrus.WithField("name", testFile.Name()).Debug("processing test file")

		createResponse, err := dockerClient.ContainerCreate(context.Background(), &containerConfig, hostConfig, nil, "")
		if err != nil {
			logrus.WithError(err).Error("Could not create container")
			return err
		}

		err = copyToContainer(context.Background(), dockerClient, testProgramReader, createResponse.ID, path.Join("/build", file.Filename))
		if err != nil {
			logrus.WithError(err).Error("Could not copy executable to container")
			return err
		}

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

		err = copyToContainer(context.Background(), dockerClient, testFileReader, createResponse.ID, path.Join("/build", "testIn"))
		if err != nil {
			logrus.WithError(err).Error("Could not copy test file to container")
			return err
		}

		if err := dockerClient.ContainerStart(context.Background(), createResponse.ID, types.ContainerStartOptions{}); err != nil {
			logrus.WithError(err).Error("Error starting container")
			return err
		}

		dockerClient.ContainerWait(context.Background(), createResponse.ID)

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

		runInstance := models.Run{
			ProjectHash: hash,
			Stdout:      logsBuffer.String()[8:],
			Stderr:      "",
		}

		db.Create(&runInstance)

		// open expected
		testFileOutputReader, err := os.Open(path.Join("projects", hash, filepath.Base(testFile.Name())+".out"))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"hash":  hash,
				"file":  path.Join("projects", hash, filepath.Base(testFile.Name())+".out"),
			}).Error("unable to open test output file while running tests")
			continue
		}
		defer testFileOutputReader.Close()

		d, err := diff(testFileOutputReader, logsBuffer)
		if err != nil {
			logrus.WithError(err).Error("unable to diff expected output with actual output")
			continue
		}

		diffs = append(diffs, d)
	}

	// TODO: Figure out why an extra eight bytes are stored at the front of the buffer.
	// \u0001 followed by seven \u0000.
	return c.JSON(http.StatusOK, diffs)
}

func copyToContainer(ctx context.Context, dockerClient *client.Client, file io.ReadCloser, dstContainer, dstPath string) (err error) {
	pipeReader, pipeWriter := io.Pipe()
	buf := bufio.NewWriterSize(pipeWriter, 32*1024)

	go func() {
		tarWriter := tar.NewWriter(buf)
		tarBuf := bufio.NewWriterSize(tarWriter, 32*1024)

		defer func() {
			buf.Flush()
			if err := tarWriter.Close(); err != nil {
				logrus.WithError(err).Error("Can't close tar writer")
			}
			if err := pipeWriter.Close(); err != nil {
				logrus.WithError(err).Error("Can't close pipe writer")
			}
		}()

		if err := addTarFile(filepath.Base(dstPath), file, tarWriter, tarBuf); err != nil {
			// if pipe is broken, stop writing tar stream to it
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

func addTarFile(filename string, file io.ReadCloser, tarWriter *tar.Writer, tarBuf *bufio.Writer) error {
	buffer := new(bytes.Buffer)
	size, err := buffer.ReadFrom(file)
	if err != nil {
		logrus.WithError(err).Error("Error reading into buffer")
		return err
	}

	hdr := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: size,
	}

	if err := tarWriter.WriteHeader(hdr); err != nil {
		logrus.WithError(err).Error("Error writing header to tar file")
		return err
	}

	_, err = io.Copy(tarBuf, buffer)
	if err != nil {
		logrus.WithError(err).Error("Error copying from buffer to tar buffer")
		return err
	}

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

func diff(in1, in2 io.Reader) (string, error) {
	// read into strings
	buf := new(bytes.Buffer)
	buf.ReadFrom(in1)
	str1 := buf.String()

	buf.Reset()
	buf.ReadFrom(in2)
	str2 := buf.String()

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(str1, str2, false)
	return dmp.DiffPrettyHtml(diffs), nil
}

func selectConfig(srcpath string) (container.Config, error) {
	var extension = filepath.Ext(srcpath)
	logrus.Debug(extension)

	timeout := constants.BUILD_TIMEOUT

	config := container.Config{
		WorkingDir:      "/build",
		NetworkDisabled: true,
		StopTimeout:     &timeout,
	}

	defaults, ok := constants.FILETYPE_CONFIGS["extension"]
	if !ok {
		logrus.WithField("extension", extension).Error("Attempting to run file with unknown extension")
		return container.Config{}, fmt.Errorf("Unsupported file extension: %s", extension)
	}

	config.Image = defaults.Image()
	config.Cmd = defaults.Command(srcpath)

	return config, nil
}
