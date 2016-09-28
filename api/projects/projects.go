package projects

import (
	"errors"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/jchorl/collabtest/constants"
	"github.com/jchorl/collabtest/models"
)

const (
	charBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charIdxBits = 6                  // 6 bits to represent a char index
	charIdxMask = 1<<charIdxBits - 1 // All 1-bits, as many as charIdxBits
	charIdxMax  = 63 / charIdxBits   // # of char indices fitting in 63 bits
)

var (
	src           = rand.NewSource(time.Now().UnixNano())
	jwtMiddleware = middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(constants.JWT_SECRET),
		TokenLookup: "cookie:Authorization",
	})
)

func Init(projects *echo.Group) {
	projects.GET("", list, jwtMiddleware)
	projects.GET("/", list, jwtMiddleware)
	projects.GET("/:hash", show)
	projects.POST("/create", create, jwtMiddleware)
	projects.DELETE("/:hash", delete, jwtMiddleware)
	projects.POST("/:hash/add", add)
	projects.POST("/:hash/run", run)
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

	hash := randomHash()
	project := models.Project{Hash: hash, Name: projectName}
	db.Create(&project)

	// create a dir for the project
	err := os.Mkdir(path.Join("projects", hash), 0700)
	if err != nil {
		logrus.WithError(err).Error("Created project but could not create project dir")
	}

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

	hash := c.Param("hash")

	project := db.Preload("Runs").Find(&models.Project{}, hash)
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

	hash := c.Param("hash")

	project := db.Find(&models.Project{}, hash)
	db.Delete(&project)
	return c.NoContent(http.StatusOK)
}

func randomHash() string {
	b := make([]byte, constants.HASH_LENGTH)
	// A src.Int63() generates 63 random bits, enough for charIdxMax characters!
	for i, cache, remain := constants.HASH_LENGTH-1, src.Int63(), charIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), charIdxMax
		}
		if idx := int(cache & charIdxMask); idx < len(charBytes) {
			b[i] = charBytes[idx]
			i--
		}
		cache >>= charIdxBits
		remain--
	}

	return string(b)
}
