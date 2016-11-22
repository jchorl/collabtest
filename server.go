package main

import (
	"context"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"

	"github.com/jchorl/collabtest/api"
	"github.com/jchorl/collabtest/constants"
	"github.com/jchorl/collabtest/models"
)

var dockerEngineHeaders = map[string]string{"User-Agent": "engine-api-cli-1.0"}

func main() {
	// Initialize logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	if os.Getenv("DEV") != "" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	// Create DB connection
	db, err := models.GetDB()
	if err != nil {
		logrus.WithError(err).Fatal("Could not connect to DB")
		return
	}

	// Create docker connection
	dockerClient, err := client.NewClient("unix:///var/run/docker.sock", "v1.22", nil, dockerEngineHeaders)
	if err != nil {
		logrus.WithError(err).Fatal("Could not create docker client")
	}

	for _, cnf := range constants.FILETYPE_CONFIGS {
		// check if we have the image
		results, err := dockerClient.ImageList(context.Background(), types.ImageListOptions{MatchName: cnf.Image()})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"image": cnf.Image(),
			}).Error("Unable to list images")
		}

		if len(results) > 0 {
			logrus.WithField("image", cnf.Image()).Debug("Already have image, skipping pull")
			continue
		}

		// Pull required docker container images to run tests
		logrus.WithField("image", cnf.Image()).Debug("About to pull image")
		_, err = dockerClient.ImagePull(context.Background(), cnf.Image(), types.ImagePullOptions{})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
				"image": cnf.Image(),
			}).Error("Unable to pull image at start")
		}
	}

	// Create router
	e := echo.New()
	// Add middleware to router
	e.Pre(middleware.HTTPSRedirect())
	e.Use(
		middleware.Logger(),
		dbMiddleware(db),
		dockerMiddleware(dockerClient),
		middleware.BodyLimit("5M"),
	)

	// Serve files
	e.File("/", "ui/build/index.html")
	e.Static("/static", "ui/build/static")

	apiRoutes := e.Group("/api")
	api.Init(apiRoutes)

	logrus.Debug("Starting server")
	// Start http server with TLS certificate
	e.Run(standard.WithTLS(":"+os.Getenv("PORT"), "/etc/letsencrypt/live/"+os.Getenv("DOMAIN")+"/fullchain.pem", "/etc/letsencrypt/live/"+os.Getenv("DOMAIN")+"/privkey.pem"))
}

func dbMiddleware(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(constants.CTX_DB, db)
			return next(c)
		}
	}
}

func dockerMiddleware(dockerClient *client.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(constants.CTX_DOCKER_CLIENT, dockerClient)
			return next(c)
		}
	}
}
