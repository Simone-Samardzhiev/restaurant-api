package repository

import (
	"context"
	"database/sql"
	"restaurant/internal/core/domain"

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
		zap.L().Error("failed to insert order", zap.Error(err))
		return domain.ErrInternal
	}

	return nil
}
