package main

import (
	// "net/http"
	"github.com/labstack/echo/v4"
	"main/handler"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	app := echo.New()
	
	app.Use(middleware.CORS())

	// server health check
	app.GET("/hc", handler.HealthCheck())

	// API v1
	apiv1 := app.Group("/api/v1")
	apiv1.GET("/", handler.GetRealIP())
	apiv1.GET("/all", handler.GetAllInfo())
	// Logger
	app.Logger.Fatal(app.Start(":8001"))

}
