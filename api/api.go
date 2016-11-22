package api

import (
	"github.com/labstack/echo"

	"github.com/jchorl/collabtest/api/auth"
	"github.com/jchorl/collabtest/api/projects"
)

// Initialize api routes
func Init(api *echo.Group) {
	projectsRoutes := api.Group("/projects")
	projects.Init(projectsRoutes)

	authRoutes := api.Group("/auth")
	auth.Init(authRoutes)
}
