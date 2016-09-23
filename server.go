package main

import (
	"github.com/jchorl/collabtest/handlers"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	e := echo.New()
	e.Use(middleware.Logger())
	e.File("/", "static/index.html")
	e.File("/new", "static/new.html")
	e.Static("/static", "static")

	api := e.Group("/api")
	handlers.Init(api)

	e.Run(standard.New(":8080"))
}
