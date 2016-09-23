package handlers

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"path"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/labstack/echo"

	"github.com/jchorl/collabtest/models"
)

const (
	BUILD_TIMEOUT = 5
)

var DEFAULT_HEADERS = map[string]string{"User-Agent": "engine-api-cli-1.0"}

func Init(api *echo.Group) {
	api.GET("/helloworld", helloWorld)
	api.POST("/create", createProject)
	api.POST("/uploadAndBuild", uploadAndBuild)
}

func uploadAndBuild(c echo.Context) error {
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

	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, DEFAULT_HEADERS)
	if err != nil {
		logrus.WithError(err).Error("Could not create docker client")
		return err
	}

	// cannot take address of const int
	timeout := BUILD_TIMEOUT

	// TODO actually select correct image
	containerConfig := &container.Config{
		Image:           "gcc",
		WorkingDir:      "/build",
		NetworkDisabled: true,
		StopTimeout:     &timeout,
		Cmd:             strslice.StrSlice{"sh", "-c", "g++ " + file.Filename + " && ./a.out"},
	}

	// TODO set ulimits in the host config
	hostConfig := &container.HostConfig{
		AutoRemove: true,
	}

	createResponse, err := cli.ContainerCreate(context.Background(), containerConfig, hostConfig, nil, "")
	if err != nil {
		logrus.WithError(err).Error("Could not create container")
		return err
	}

	err = copyToContainer(context.Background(), cli, src, createResponse.ID, path.Join("/build", file.Filename))
	if err != nil {
		logrus.WithError(err).Error("Could not copy to container")
		return err
	}

	if err := cli.ContainerStart(context.Background(), createResponse.ID, types.ContainerStartOptions{}); err != nil {
		logrus.WithError(err).Error("Error starting container")
		return err
	}

	// TODO wait until container is complete

	return c.NoContent(http.StatusAccepted)
}

func copyToContainer(ctx context.Context, cli *client.Client, file io.ReadCloser, dstContainer, dstPath string) (err error) {
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

	return cli.CopyToContainer(ctx, dstContainer, filepath.Dir(dstPath), pipeReader, types.CopyToContainerOptions{})
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

func helloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func createProject(c echo.Context) error {
	projectName := c.FormValue("name")
	db, err := models.GetDB()
	if err != nil {
		return err
	}

	project := models.Project{Name: projectName}
	db.Create(&project)

	return c.String(http.StatusOK, "Connected to postgres")
}
