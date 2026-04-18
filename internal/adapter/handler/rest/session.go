package rest

import (
	"net/http"
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionHandler handles session related http requests.
type SessionHandler struct {
	service order.SessionService
}

// NewSessionHandler allocates and creates new [SessionHandler].
func NewSessionHandler(service order.SessionService) *SessionHandler {
	return &SessionHandler{
		service: service,
	}
}

// SessionResponse represents the JSON response of a session.
type SessionResponse struct {
	Id     uuid.UUID `json:"id"`
	Table  int       `json:"table"`
	Status string    `json:"status"`
}

// OrderedProductResponse represents the JSON response of an ordered product.
type OrderedProductResponse struct {
	Id        uuid.UUID `json:"id"`
	ProductId uuid.UUID `json:"productId"`
	SessionId uuid.UUID `json:"sessionId"`
	Status    string    `json:"status"`
}

// SessionDetailsResponse represents the JSON response of session details.
type SessionDetailsResponse struct {
	SessionResponse
	OrderedProducts []OrderedProductResponse `json:"orderedProducts"`
}

func (h *SessionHandler) GetSessionDetails(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		ctx.Error(domain.NewBadRequestError("invalid uuid", domain.ErrorCodeInvalidUUID, err)).SetType(gin.ErrorTypePublic)
		return
	}

	result, err := h.service.GetSessionDetails(ctx, id)
	if err != nil {
		ctx.Error(err).SetType(gin.ErrorTypePublic)
		return
	}

	response := SessionDetailsResponse{
		SessionResponse: SessionResponse{
			Id:     id,
			Table:  result.Table.Number(),
			Status: result.Status.String(),
		},
		OrderedProducts: make([]OrderedProductResponse, 0, len(result.OrderedProducts)),
	}

	for _, orderedProduct := range result.OrderedProducts {
		response.OrderedProducts = append(response.OrderedProducts, OrderedProductResponse{
			Id:        orderedProduct.Id,
			ProductId: orderedProduct.ProductId,
			SessionId: orderedProduct.SessionId,
			Status:    orderedProduct.Status.String(),
		})
	}

	ctx.JSON(http.StatusOK, response)
}
