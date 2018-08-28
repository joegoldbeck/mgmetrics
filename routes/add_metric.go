package routes

import (
	"github.com/joegoldbeck/mgmetrics/models"
	"github.com/labstack/echo"
	"net/http"
	"time"
)

// AddMetricHandler is the handler for the POST /api/metrics route
// It returns the metric inserted upon success
func AddMetricHandler(db *models.DB) func(echo.Context) error {
	return func(c echo.Context) (err error) {
		incomingMetric := new(models.IncomingMetric)
		if err = c.Bind(incomingMetric); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err = incomingMetric.Validate(); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		metric := models.PlainMetric{Timestamp: time.Now().UnixNano() / 1000000, Key: incomingMetric.Key, Value: incomingMetric.Value, Tags: incomingMetric.Tags}
		if _, err = db.AddMetric(metric); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.JSON(http.StatusCreated, metric)
	}
}
