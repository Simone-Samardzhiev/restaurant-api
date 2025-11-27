package service

import (
	"context"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port"

	"github.com/google/uuid"
)

// OrderService implements port.OrderService and provided access to orders-related business logic
type OrderService struct {
	orderRepository port.OrderRepository
}

// NewOrderService creates new OrderService interface.
func NewOrderService(orderRepository port.OrderRepository) *OrderService {
	return &OrderService{
		orderRepository: orderRepository,
	}
}

func (s *OrderService) GetSessions(ctx context.Context) ([]domain.OrderSession, error) {
	return s.orderRepository.GetSessions(ctx)
}

func (s *OrderService) CreateSession(ctx context.Context) (*domain.OrderSession, error) {
	order := domain.NewSession(uuid.New(), 1, domain.Closed)
	err := s.orderRepository.AddSession(ctx, order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) UpdateSession(ctx context.Context, session *domain.UpdateOrderSessionDTO) error {
	hasUpdate := false
	switch {
	case session.NewTableNumber != nil:
		hasUpdate = true
	case session.NewStatus != nil:
		hasUpdate = true
	}

	if !hasUpdate {
		return domain.ErrNothingToUpdate
	}

	return s.orderRepository.UpdateSession(ctx, session)
}

func (s *OrderService) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.orderRepository.DeleteSession(ctx, id)
}

func (s *OrderService) ValidateSession(ctx context.Context, sessionId uuid.UUID) error {
	session, err := s.orderRepository.GetSessionByID(ctx, sessionId)
	if err != nil {
		return err
	}

	if session.Status != domain.Open {
		return domain.ErrOrderSessionIsNotOpen
	}
	return nil
}
