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
	projects.POST("/create", create)
	projects.POST("/submit", submit)
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
