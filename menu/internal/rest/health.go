package rest

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler has a handler function to check if the service is healthy.
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates and allocates new [HealthHandler].
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// UnhealthyResponse represents the JSON error response if the service is unhealthy.
type UnhealthyResponse struct {
	Cause string `json:"cause"`
}

func (h *HealthHandler) IsHealthy(ctx *gin.Context) {
	if err := h.db.Ping(); err != nil {
		ctx.Status(http.StatusOK)
	} else {
		ctx.JSON(http.StatusServiceUnavailable, UnhealthyResponse{
			Cause: "Database connection is lost.",
		})
	}
}
