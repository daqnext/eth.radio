package api

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
)

func testapi(c echo.Context) error {
	return c.String(http.StatusOK, "hello word")
}

func DeclareRoute(server *echo.Echo) {

	server.GET("/test/api", testapi)
}
