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
	"path"
	"path/filepath"
	"strconv"

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

func submit(c echo.Context) error {
	dockerClient, ok := c.Get(constants.CTX_DOCKER_CLIENT).(*client.Client)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get docker client from context in project submit")
		return errors.New("Unable to get docker client from context")
	}

	file, err := c.FormFile("file")
	if err != nil {
		logrus.WithError(err).Error("Unable to get file from upload req")
		return err
	}

	src, err := file.Open()
	if err != nil {
		logrus.WithError(err).Error("Could not open uploaded file to write to disk")
		return err
	}
	defer src.Close()

	containerConfig, err := selectConfig(file.Filename)
	if err != nil {
		return err
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Resources: container.Resources{
			CPUShares: constants.BUILD_CPU_SHARE,
			Memory:    constants.BUILD_MEMORY,
		},
	}

	createResponse, err := dockerClient.ContainerCreate(context.Background(), &containerConfig, hostConfig, nil, "")
	if err != nil {
		logrus.WithError(err).Error("Could not create container")
		return err
	}

	err = copyToContainer(context.Background(), dockerClient, src, createResponse.ID, path.Join("/build", file.Filename))
	if err != nil {
		logrus.WithError(err).Error("Could not copy to container")
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

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id": c.Param("id"),
		}).Error("Unable to convert id from str to int in project submit")
	}
	submission := models.Submission{
		ProjectID: uint(id),
		Stdout:    logsBuffer.String()[8:],
		Stderr:    "",
	}
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project submit")
		return errors.New("Unable to get DB from context")
	}

	db.Create(&submission)

	// TODO: Figure out why an extra eight bytes are stored at the front of the buffer.
	// \u0001 followed by seven \u0000.
	return c.String(http.StatusOK, logsBuffer.String()[8:])
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

func diff(c echo.Context) error {
	// Hardcoded stub. Normally, we would generate the output by running the program on each of the inputs.
	diff, err := fileDiff("projects/incorrect/output/test.out", "projects/incorrect/expected/test.exp")
	if err != nil {
		logrus.WithError(err).Error("Error computing file diff.")
		return err
	}
	logrus.Debug(diff)
	return c.HTML(http.StatusOK, diff)
}

func fileDiff(outpath string, exppath string) (string, error) {
	outbuf, err := ioutil.ReadFile(outpath)
	if err != nil {
		logrus.WithError(err).Error("Error opening output file for reading.")
		return "", err
	}
	out := string(outbuf)

	expbuf, err := ioutil.ReadFile(exppath)
	if err != nil {
		logrus.WithError(err).Error("Error opening expected output file for reading.")
		return "", err
	}
	exp := string(expbuf)

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(out, exp, false)
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
