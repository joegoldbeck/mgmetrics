package main

import (
	"github.com/facebookgo/grace/gracehttp"
	"github.com/joegoldbeck/mgmetrics/models"
	"github.com/joegoldbeck/mgmetrics/routes"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Database
	// In production, these would obviously be stored in a secrets store and passed in via application configuration.
	db, err := models.OpenDB("host=localhost port=5432 user=metrics_user dbname=postgres password=dev_only sslmode=disable")

	if err != nil {
		e.Logger.Fatal("Error connecting to database: ", err)
	}

	defer db.Close()

	// Routes
	// This just serves a nice server-is-up message
	e.GET("/", routes.HomeHandler())
	// curl -X GET http://localhost:8080/api/metrics\?tag\="bed-22"\&key\="heartrate"
	// curl -X GET http://localhost:8080/api/metrics\?&key\="heartrate"&minTimestamp\=1505879975574&maxTimestamp\=1505979975574
	e.GET("/api/metrics", routes.GetMetricsHandler(db))
	// curl -H "Content-Type: application/json" -X POST -d '{"key": "heartrate", "value": 52.4, "tags": ["icu", "ward-4", "bed-22"]}' http://localhost:8080/api/metrics
	e.POST("/api/metrics", routes.AddMetricHandler(db))

	// In production, this would come from configuration.
	e.Server.Addr = ":8080"

	// Start the server
	e.Logger.Fatal(gracehttp.Serve(e.Server))
}
