package rest

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
)

var errEndpointNotFound = errors.New("endpoint not found")

func handleEndpointNotFound(c *echo.Context) error {
	return c.JSON(http.StatusNotFound, &ErrorResponse{
		HTTPStatus: http.StatusNotFound,
		Message:    "Endpoint not found.",
		ErrorCode:  "ENDPOINT_NOT_FOUND",
		err:        errEndpointNotFound,
	})
}
