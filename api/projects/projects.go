package projects

import (
	"errors"
	"math/rand"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
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
	projects.POST("", create, jwtMiddleware)
	projects.GET("/:hash", show)
	projects.DELETE("/:hash", delete, jwtMiddleware)
	projects.POST("/:hash/add", add)
	projects.POST("/:hash/run", run)
}

func create(c echo.Context) error {
	project := models.Project{}
	if err := c.Bind(&project); err != nil {
		logrus.WithFields(logrus.Fields{
			"context": c,
			"error":   err,
		}).Error("Unable to decode project in project create")
	}

	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project create")
		return errors.New("Unable to get DB from context")
	}

	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get user from context in project list")
		return errors.New("Unable to get user from context")
	}

	claims := user.Claims.(jwt.MapClaims)
	userIdF := claims["sub"].(float64)
	userId := uint(userIdF)

	hash := randomHash()
	project.Hash = hash
	project.UserId = userId
	result := db.Create(&project)
	if result.Error != nil {
		logrus.WithFields(logrus.Fields{
			"error":   result.Error,
			"project": project,
		}).Error("Unable to insert new project into db")
		return result.Error
	}

	// create a dir for the project
	// dont fail on this, try again on file upload and fail on that
	err := os.Mkdir(path.Join("projects", hash), 0700)
	if err != nil {
		logrus.WithError(err).Error("Created project but could not create project dir")
	}

	return c.JSON(http.StatusOK, project)
}

func list(c echo.Context) error {
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project list")
		return errors.New("Unable to get DB from context")
	}

	user, ok := c.Get("user").(*jwt.Token)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get user from context in project list")
		return errors.New("Unable to get user from context")
	}

	claims := user.Claims.(jwt.MapClaims)
	userIdF := claims["sub"].(float64)
	userId := uint(userIdF)

	result := db.Where("user_id = ?", userId).Find(&[]models.Project{})
	if result.Error != nil {
		logrus.WithError(result.Error).Error("Error querying for projects in project list")
		return result.Error
	}
	return c.JSON(http.StatusOK, result.Value)
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
