package order

import (
	"context"

	"github.com/google/uuid"
)

// DefaultSessionService is the default implementation of [SessionService].
type DefaultSessionService struct {
	sessionRepository        SessionRepository
	orderedProductRepository OrderedProductRepository
}

var _ SessionService = (*DefaultSessionService)(nil)

// NewDefaultSessionService allocates and creates a new [DefaultSessionService].
func NewDefaultSessionService(sessionRepository SessionRepository, orderedProductRepository OrderedProductRepository) *DefaultSessionService {
	return &DefaultSessionService{
		sessionRepository:        sessionRepository,
		orderedProductRepository: orderedProductRepository,
	}
}

func (s *DefaultSessionService) AddSession(ctx context.Context, request *AddSessionRequest) (*Session, error) {
	return s.sessionRepository.Save(ctx, request)
}

func (s *DefaultSessionService) UpdateSession(ctx context.Context, request *UpdateSessionRequest) error {
	return s.sessionRepository.Update(ctx, request)
}

func (s *DefaultSessionService) GetSessionDetails(ctx context.Context, id uuid.UUID) (*SessionDetails, error) {
	session, err := s.sessionRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	products, err := s.orderedProductRepository.GetBySessionId(ctx, id)
	if err != nil {
		return nil, err
	}

	return &SessionDetails{
		Session:         *session,
		OrderedProducts: products,
	}, nil
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

func (s *DefaultOrderedProductService) DeleteOrder(ctx context.Context, id uuid.UUID) (*OrderedProduct, error) {
	return s.repository.Delete(ctx, id)
}
