package rest

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/valkey-io/valkey-go"
)

// HealthHandler has a handler function to check if the service is healthy.
type HealthHandler struct {
	db         *sql.DB
	valkeyConn valkey.Client
}

// NewHealthHandler creates and allocates new [HealthHandler].
func NewHealthHandler(db *sql.DB, valkeyConn valkey.Client) *HealthHandler {
	return &HealthHandler{db: db, valkeyConn: valkeyConn}
}

// UnhealthyResponse represents the JSON error response if the service is unhealthy.
type UnhealthyResponse struct {
	Cause string `json:"cause"`
}

func (h *HealthHandler) IsHealthy(ctx *echo.Context) error {
	if err := h.db.Ping(); err != nil {
		return ctx.JSON(http.StatusServiceUnavailable, UnhealthyResponse{
			Cause: "Database connection is lost.",
		})
	}

	if err := h.valkeyConn.Do(ctx.Request().Context(), h.valkeyConn.B().Ping().Build()).Error(); err != nil {
		return ctx.JSON(http.StatusServiceUnavailable, UnhealthyResponse{
			Cause: "Valkey connection is lost.",
		})
	}

	return ctx.NoContent(http.StatusOK)
}
