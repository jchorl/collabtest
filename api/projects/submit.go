package projects

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"path"
	"path/filepath"

    "github.com/sergi/go-diff/diffmatchpatch"
	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/labstack/echo"

	"github.com/jchorl/collabtest/constants"
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

	// cannot take address of const int
	timeout := constants.BUILD_TIMEOUT

	// TODO actually select correct image
	containerConfig := &container.Config{
		Image:           "frolvlad/alpine-gcc",
		WorkingDir:      "/build",
		NetworkDisabled: true,
		StopTimeout:     &timeout,
		Cmd:             strslice.StrSlice{"sh", "-c", "g++ " + file.Filename + " && ./a.out"},
	}

	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Resources: container.Resources{
			CPUShares: constants.BUILD_CPU_SHARE,
			Memory:    constants.BUILD_MEMORY,
		},
	}

	createResponse, err := dockerClient.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, "")
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

	return c.String(http.StatusOK, logsBuffer.String())
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

func fileDiff(outfile string, expfile string) error {
    return nil
}
