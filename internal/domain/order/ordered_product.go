package order

import (
	"errors"
	"restaurant/internal/domain"

	"github.com/google/uuid"
)

var (
	StatusPending   = OrderedProductStatus{"pending"}
	StatusPreparing = OrderedProductStatus{"preparing"}
	StatusDone      = OrderedProductStatus{"done"}
)

// OrderedProductStatus represents a valid ordered product status.
type OrderedProductStatus struct {
	raw string
}

func (s *OrderedProductStatus) String() string {
	return s.raw
}

func (s *OrderedProductStatus) Equals(status OrderedProductStatus) bool {
	return s.raw == status.raw
}

// ParseOrderedProductStatus parses [OrderedProductStatus] from string.
//
// If the status is invalid the error will be of type [domain.ErrorDetail]
func ParseOrderedProductStatus(status string) (OrderedProductStatus, error) {
	switch status {
	case StatusPending.raw, StatusPreparing.raw, StatusDone.raw:
		return OrderedProductStatus{status}, nil
	default:
		return OrderedProductStatus{}, &domain.ErrorDetail{
			Code:    domain.ErrorCodeInvalidOrderedProductStatus,
			Message: "invalid ordered product status",
			Metadata: map[string]any{
				"actual": status,
				"supported": []string{
					StatusPending.raw,
					StatusPreparing.raw,
					StatusDone.raw,
				},
			},
		}
	}
}

// OrderedProduct represents a valid ordered product entity.
type OrderedProduct struct {
	Id        uuid.UUID
	ProductId uuid.UUID
	SessionId uuid.UUID
	Status    OrderedProductStatus
}

// ParseOrderedProduct parses [OrderedProduct] from id, product id, session id and status.
//
// If the product is invalid the error will be of type [domain.Error].
func ParseOrderedProduct(id, productId, sessionId uuid.UUID, status string) (*OrderedProduct, error) {
	errs := make([]domain.ErrorDetail, 0)
	parsedStatus, err := ParseOrderedProductStatus(status)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *detailErr)
		} else {
			return nil, err
		}
	}

	if len(errs) > 0 {
		return nil, domain.NewValidationError("invalid ordered product", domain.ErrorCodeInvalidOrderedProduct, errs...)
	}

	return &OrderedProduct{
		Id:        id,
		ProductId: productId,
		SessionId: sessionId,
		Status:    parsedStatus,
	}, nil
}

// AddOrderedProductRequest represents a request for adding a new ordered product.
type AddOrderedProductRequest struct {
	ProductId uuid.UUID
	SessionId uuid.UUID
}

// NewAddOrderedProductRequest allocates and creates new [AddOrderedProductRequest].
func NewAddOrderedProductRequest(productId uuid.UUID, sessionId uuid.UUID) *AddOrderedProductRequest {
	return &AddOrderedProductRequest{
		ProductId: productId,
		SessionId: sessionId,
	}
}
