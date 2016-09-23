package api

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/jchorl/collabtest/api/projects"
)

func Init(api *echo.Group) {
	projectsRoutes := api.Group("/projects")
	projects.Init(projectsRoutes)

	api.GET("/helloworld", helloWorld)
}

func helloWorld(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
