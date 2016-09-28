package projects

import (
	"errors"
	"net/http"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"

	"github.com/jchorl/collabtest/constants"
	"github.com/jchorl/collabtest/models"
)

func Init(projects *echo.Group) {
	projects.GET("", list)
	projects.GET("/", list)
	projects.GET("/:id", show)
	projects.POST("/create", create)
	projects.DELETE("/:id", delete)
	projects.POST("/submit", submit)
	projects.GET("/diff", diff)
}

func create(c echo.Context) error {
	projectName := c.FormValue("name")
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project create")
		return errors.New("Unable to get DB from context")
	}

	project := models.Project{Name: projectName}
	db.Create(&project)

	return c.String(http.StatusOK, "Created project: "+projectName)
}

func list(c echo.Context) error {
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project list")
		return errors.New("Unable to get DB from context")
	}

	projects := db.Find(&[]models.Project{})
	return c.JSON(http.StatusOK, projects)
}

func show(c echo.Context) error {
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project show")
		return errors.New("Unable to get DB from context")
	}

	id := c.Param("id")

	project := db.Find(&models.Project{}, id)
	return c.JSON(http.StatusOK, project)
}

func delete(c echo.Context) error {
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project delete")
		return errors.New("Unable to get DB from context")
	}

	id := c.Param("id")

	project := db.Find(&models.Project{}, id)
	db.Delete(&project)
	return c.NoContent(http.StatusOK)
}
