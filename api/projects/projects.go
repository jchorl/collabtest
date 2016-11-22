package projects

import (
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	"github.com/jchorl/collabtest/constants"
	"github.com/jchorl/collabtest/models"
)

// Contants for hash generation
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

type Testcase struct {
	InputLink  string `json:"inputLink"`
	OutputLink string `json:"outputLink"`
}

// Initialize routes for projects
func Init(projects *echo.Group) {
	projects.GET("", list, jwtMiddleware)
	projects.POST("", create, jwtMiddleware)
	projects.GET("/:hash", show)
	projects.GET("/:hash/testcases", listTestcases)
	projects.GET("/:hash/testcases/:filename", getTestcase)
	projects.DELETE("/:hash", deleteProject, jwtMiddleware)
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

	// Get DB connection
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project create")
		return errors.New("Unable to get DB from context")
	}

	// Get user JWT parsed by middleware
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

	// Create project for user
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

	// Create a dir for the project
	// Dont fail on this, try again on file upload and fail on that
	err := os.Mkdir(path.Join("projects", hash), 0700)
	if err != nil {
		logrus.WithError(err).Error("Created project but could not create project dir")
	}

	return c.JSON(http.StatusOK, project)
}

// Listing a users project
func list(c echo.Context) error {
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project list")
		return errors.New("Unable to get DB from context")
	}

	// Get user JWT parsed by middleware
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

// List test cases of a project
func listTestcases(c echo.Context) error {
	hash := c.Param("hash")

	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project list")
		return errors.New("Unable to get DB from context")
	}

	// Verify that the project exists
	if db.First(&models.Project{Hash: hash}).RecordNotFound() {
		return constants.UNRECOGNIZED_HASH
	}

	testFiles, err := ioutil.ReadDir(path.Join("projects", hash))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"dir":   path.Join("projects", hash),
		}).Error("unable to ls dir to get testcases")
	}

	testcases := map[string]Testcase{}
	for _, testFile := range testFiles {
		testcaseHash := testFile.Name()[0 : len(testFile.Name())-len(filepath.Ext(testFile.Name()))] // disgusting way of removing extension. TODO improve.

		testcase := testcases[testcaseHash]
		if strings.HasSuffix(testFile.Name(), ".in") {
			testcase.InputLink = getTestcaseLink(hash, testFile.Name())
		} else if strings.HasSuffix(testFile.Name(), ".out") {
			testcase.OutputLink = getTestcaseLink(hash, testFile.Name())
		}

		testcases[testcaseHash] = testcase
	}

	return c.JSON(http.StatusOK, testcases)
}

func getTestcaseLink(hash, filename string) string {
	return "/api/projects/" + hash + "/testcases/" + filename
}

// Return test file contents
func getTestcase(c echo.Context) error {
	hash := c.Param("hash")
	filename := c.Param("filename")

	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project list")
		return errors.New("Unable to get DB from context")
	}

	// Verify that the project exists
	if db.First(&models.Project{Hash: hash}).RecordNotFound() {
		return constants.UNRECOGNIZED_HASH
	}

	file, err := os.Open(path.Join("projects", hash, filename))
	if err != nil {
		return err
	}

	// Add .txt so the browser can display the file, and so the client's computer knows how to open it
	return c.Inline(file, filename+".txt")
}

// Return info about a specific project
func show(c echo.Context) error {
	db, ok := c.Get(constants.CTX_DB).(*gorm.DB)
	if !ok {
		logrus.WithFields(logrus.Fields{
			"context": c,
		}).Error("Unable to get DB from context in project show")
		return errors.New("Unable to get DB from context")
	}

	hash := c.Param("hash")

	result := db.First(&models.Project{Hash: hash})
	return c.JSON(http.StatusOK, result.Value)
}

// Delete a project
func deleteProject(c echo.Context) error {
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

// Help for generating random hash for project
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
