package handlers

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/labstack/echo"
)

func Init(api *echo.Group) {
	api.GET("/helloworld", helloWorld)
	api.POST("/file", fileUpload)
	api.GET("/dockerps", dockerPs)
}

func helloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}

func fileUpload(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// TODO store file in correct application dir (either it's a test case or an actual application)
	dst, err := os.Create(file.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func dockerPs(c echo.Context) error {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	cli, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, defaultHeaders)
	if err != nil {
		return err
	}

	options := types.ContainerListOptions{All: true}
	containers, err := cli.ContainerList(context.Background(), options)
	if err != nil {
		return err
	}

	ctrs := [][]string{}
	for _, ctr := range containers {
		ctrs = append(ctrs, ctr.Names)
	}

	return c.JSON(http.StatusOK, ctrs)
}
