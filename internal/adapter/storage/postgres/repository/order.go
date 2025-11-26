package repository

import (
	"context"
	"database/sql"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OrderRepository implements port.OrderRepository and provides access to postgres.
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository creates a new OrderRepository instance.
func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

func (r *OrderRepository) GetSessions(ctx context.Context) ([]domain.OrderSession, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, table_number, status FROM order_sessions")
	if err != nil {
		zap.L().Error("error getting product", zap.Error(err))
		return nil, domain.ErrInternal
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			zap.L().Warn("error closing rows", zap.Error(closeErr))
		}
	}()

	var sessions []domain.OrderSession
	for rows.Next() {
		var session domain.OrderSession
		if err = rows.Scan(&session.Id, &session.TableNumber, &session.Status); err != nil {
			zap.L().Error("error scanning row", zap.Error(err))
			return nil, domain.ErrInternal
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func (r *OrderRepository) AddSession(ctx context.Context, order *domain.OrderSession) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO order_sessions(id, table_number, status) 
		VALUES ($1, $2, $3)`,
		order.Id,
		order.TableNumber,
		order.Status,
	)

	if err != nil {
		zap.L().Error("error inserting order", zap.Error(err))
		return domain.ErrInternal
	}

	return nil
}

func (r *OrderRepository) DeleteSession(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM order_sessions WHERE id = $1", id)
	if err != nil {
		zap.L().Error(
			"error deleting order_session",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		return domain.ErrInternal
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zap.L().Error("error getting rows affected", zap.Error(err))
		return domain.ErrInternal
	}

	if rows == 0 {
		return domain.ErrOrderSessionNotFound
	}
	return nil
}
