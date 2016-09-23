package projects

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/jchorl/collabtest/models"
)

func Init(projects *echo.Group) {
	projects.POST("/create", create)
	projects.POST("/submit", submit)
}

func create(c echo.Context) error {
	projectName := c.FormValue("name")
	db, err := models.GetDB()
	if err != nil {
		return err
	}

	project := models.Project{Name: projectName}
	db.Create(&project)

	return c.String(http.StatusOK, "Created project: "+projectName)
}
