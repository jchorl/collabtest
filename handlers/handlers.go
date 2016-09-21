package handlers

import (
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo"
)

func Init(api *echo.Group) {
	api.GET("/helloworld", helloWorld)
	api.POST("/file", fileUpload)
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
