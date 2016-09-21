package main

import (
	"github.com/jchorl/collabtest/handlers"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

func main() {
	e := echo.New()
	e.GET("/", handlers.HelloWorld)
	e.Run(standard.New(":8080"))
}
