package routes

import (
	"github.com/labstack/echo"
	"net/http"
)

// HomeHandler is the handler for the GET / route
// It returns a nice message
func HomeHandler() func(echo.Context) error {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome. The server is up and running!")
	}
}
