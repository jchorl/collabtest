package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"

	"github.com/jchorl/collabtest/api"
	"github.com/jchorl/collabtest/constants"
	"github.com/jchorl/collabtest/models"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	db, err := models.GetDB()
	if err != nil {
		logrus.WithError(err).Fatal("Could not connect to DB")
		return
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(dbMiddleware(db))
	e.File("/", "static/index.html")
	e.File("/new", "static/new.html")
	e.Static("/static", "static")

	apiRoutes := e.Group("/api")
	api.Init(apiRoutes)

	e.Run(standard.New(":8080"))
}

func dbMiddleware(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(constants.CTX_DB, db)
			return next(c)
		}
	}
}
