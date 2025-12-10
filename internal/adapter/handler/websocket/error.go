package websocket

import (
	"errors"
	"restaurant/internal/core/domain"

	"github.com/gofiber/websocket/v2"
	"go.uber.org/zap"
)

// isExpectedCloseError checks if the websocket close error is expected.
func isExpectedCloseError(err error) bool {
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived)
}

// handleDomainError handle domain error and logs any unexpected errors.
func handleDomainError(conn *websocket.Conn, err error) {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		writeString("Product not found", conn)

	case errors.Is(err, domain.ErrOrderSessionNotFound):
		writeString("Session not found", conn)

	case errors.Is(err, domain.ErrOrderSessionIsNotOpen):
		writeString("Session is not open", conn)

	case errors.Is(err, domain.ErrOrderedProductNotFound):
		writeString("Ordered product not found", conn)

	case errors.Is(err, domain.ErrOrderedProductNotPending):
		writeString("Only pending products can be deleted by a client", conn)

	case errors.Is(err, domain.ErrNothingToUpdate):
		writeString("Nothing to update", conn)
	default:
		zap.L().Error("Unknown error", zap.Error(err))
		writeString("Internal server error", conn)
	}
}
