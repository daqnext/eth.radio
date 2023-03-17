package api

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
)

func health(c echo.Context) error {
	return c.String(http.StatusOK, "it's health")
}

func DeclareRoute(server *echo.Echo) {
	server.GET("/health", health)
}
