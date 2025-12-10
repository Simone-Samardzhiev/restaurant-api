package repository

import (
	"context"
	"database/sql"
	"errors"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

func (r *OrderRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.OrderSession, error) {
	row := r.db.QueryRowContext(
		ctx,
		"SELECT id, table_number, status FROM order_sessions WHERE id = $1",
		id,
	)

	var session domain.OrderSession
	err := row.Scan(&session.Id, &session.TableNumber, &session.Status)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderSessionNotFound
	} else if err != nil {
		zap.L().Error("error scanning row", zap.Error(err))
		return nil, domain.ErrInternal
	}

	return &session, nil
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

func (r *OrderRepository) UpdateSession(ctx context.Context, session *domain.UpdateOrderSessionDTO) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE order_sessions
		SET table_number = COALESCE($1, table_number),
    		status       = COALESCE($2, status)
		WHERE id = $3`,
		session.NewTableNumber,
		session.NewStatus,
		session.Id,
	)
	if err != nil {
		zap.L().Error("error updating order", zap.Error(err))
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

func (r *OrderRepository) AddOrderedProduct(ctx context.Context, product *domain.OrderedProduct) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO ordered_products(id, product_id, session_id, status) VALUES ($1, $2, $3, $4)`,
		product.Id,
		product.ProductId,
		product.OrderSessionID,
		product.Status,
	)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23503" && pqErr.Constraint == "ordered_products_product_id_fkey" {
			return domain.ErrProductNotFound
		}
	} else if err != nil {
		zap.L().Error("error inserting ordered product", zap.Error(err))
		return domain.ErrInternal
	}

	return nil
}

func (r *OrderRepository) DeletePendingOrderedProduct(ctx context.Context, orderedProductId uuid.UUID) (*domain.OrderedProduct, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		zap.L().Error("error starting transaction", zap.Error(err))
		return nil, domain.ErrInternal
	}

	row := tx.QueryRowContext(
		ctx, `DELETE FROM ordered_products 
       	WHERE id = $1
       	RETURNING id, product_id, session_id, status`,
		orderedProductId,
	)

	var orderedProduct domain.OrderedProduct
	err = row.Scan(
		&orderedProduct.Id,
		&orderedProduct.ProductId,
		&orderedProduct.OrderSessionID,
		&orderedProduct.Status,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderedProductNotFound
	} else if err != nil {
		zap.L().Error("error scanning row", zap.Error(err))
		return nil, domain.ErrInternal
	}

	if orderedProduct.Status != domain.Pending {
		err = tx.Rollback()
		if err != nil {
			zap.L().Warn("error rolling back transaction", zap.Error(err))
		}
		return nil, domain.ErrOrderedProductNotPending
	}

	err = tx.Commit()
	if err != nil {
		zap.L().Warn("error committing transaction", zap.Error(err))
	}

	return &orderedProduct, nil
}

func (r *OrderRepository) DeleteOrderedProduct(ctx context.Context, orderedProductId uuid.UUID) (*domain.OrderedProduct, error) {
	row := r.db.QueryRowContext(
		ctx, `DELETE FROM ordered_products 
       	WHERE id = $1
       	RETURNING id, product_id, session_id, status`,
		orderedProductId,
	)

	var orderedProduct domain.OrderedProduct
	err := row.Scan(
		&orderedProduct.Id,
		&orderedProduct.ProductId,
		&orderedProduct.OrderSessionID,
		&orderedProduct.Status,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderedProductNotFound
	} else if err != nil {
		zap.L().Error("error scanning row", zap.Error(err))
		return nil, domain.ErrInternal
	}

	return &orderedProduct, nil
}
