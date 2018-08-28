package routes

import (
	"github.com/joegoldbeck/mgmetrics/models"
	"github.com/labstack/echo"
	"net/http"
)

// GetMetricsHandler is the handler for the GET /api/metrics route.
// It returns the metrics matching the query upon success.
// This is fairly simple for now. In production, we might want
// something that aided in paging, like an indication of whether
// more results were available
func GetMetricsHandler(db *models.DB) func(echo.Context) error {
	return func(c echo.Context) (err error) {
		opts := new(models.GetMetricsOpts)
		if err = c.Bind(opts); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		metrics, err := db.GetMetrics(*opts)
		if err != nil {
			return
		}
		return c.JSON(http.StatusOK, metrics)
	}
}
