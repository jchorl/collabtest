package handlers

import (
	"net/http"

	"github.com/labstack/echo"
)

func Init(api *echo.Group) {
	api.GET("/helloworld", helloWorld)
}

func helloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
