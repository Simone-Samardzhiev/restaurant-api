package order

import (
	"context"
)

// DefaultSessionService is the default implementation of [SessionService].
type DefaultSessionService struct {
	repository SessionRepository
}

var _ SessionService = (*DefaultSessionService)(nil)

// NewDefaultSessionService allocates and creates a new [DefaultSessionService].
func NewDefaultSessionService(repository SessionRepository) *DefaultSessionService {
	return &DefaultSessionService{
		repository,
	}
}

func (s *DefaultSessionService) AddSession(ctx context.Context, request *AddSessionRequest) (*Session, error) {
	return s.repository.Save(ctx, request)
}

func (s *DefaultSessionService) UpdateSession(ctx context.Context, request *UpdateSessionRequest) error {
	return s.repository.Update(ctx, request)
}

// DefaultOrderedProductService is the default implementation of [OrderedProductService].
type DefaultOrderedProductService struct {
	repository OrderedProductRepository
}

var _ OrderedProductService = (*DefaultOrderedProductService)(nil)

// NewDefaultOrderedProductService allocates and creates new [DefaultOrderedProductService].
func NewDefaultOrderedProductService(repository OrderedProductRepository) *DefaultOrderedProductService {
	return &DefaultOrderedProductService{
		repository,
	}
}

func (s *DefaultOrderedProductService) PlaceOrder(ctx context.Context, request *AddOrderedProductRequest) (*OrderedProduct, error) {
	return s.repository.Save(ctx, request)
}
