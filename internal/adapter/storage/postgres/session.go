package postgres

import (
	"database/sql"
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"

	"context"

	"github.com/google/uuid"
)

// SessionRepository implements [order.SessionRepository] using postgres.
type SessionRepository struct {
	db *sql.DB
}

var _ order.SessionRepository = (*SessionRepository)(nil)

// NewSessionRepository allocates and creates a new [SessionRepository].
func NewSessionRepository(db *sql.DB) SessionRepository {
	return SessionRepository{
		db: db,
	}
}

func (r *SessionRepository) Save(ctx context.Context, request *order.AddSessionRequest) (*order.Session, error) {
	session := &order.Session{
		Id:     uuid.New(),
		Table:  request.Table,
		Status: request.Status,
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO order_sessions(id, table_number, status) VALUES ($1, $2, $3)`,
		session.Id,
		session.Table.Number(),
		session.Status.String(),
	)

	if err != nil {
		return nil, domain.NewInternalError("error inserting session", err)
	}

	return session, nil
}
