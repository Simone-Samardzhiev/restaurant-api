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
