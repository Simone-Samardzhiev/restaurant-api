package repository

import (
	"context"
	"database/sql"
	"errors"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
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

func (r *OrderRepository) UpdateSession(ctx context.Context, session *domain.UpdateOrderSessionDTO) (*domain.OrderSession, error) {
	row := r.db.QueryRowContext(
		ctx,
		`UPDATE order_sessions
		SET table_number = COALESCE($1, table_number),
    		status       = COALESCE($2, status)
		WHERE id = $3
		RETURNING id, table_number, status`,
		session.NewTableNumber,
		session.NewStatus,
		session.Id,
	)

	var orderSession domain.OrderSession
	err := row.Scan(&orderSession.Id, &orderSession.TableNumber, &orderSession.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrOrderSessionNotFound
	} else if err != nil {
		zap.L().Error("error scanning row", zap.Error(err))
		return nil, domain.ErrInternal
	}

	return &orderSession, nil
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

func (r *OrderRepository) UpdateOrderedProductStatus(ctx context.Context, id uuid.UUID, status domain.OrderedProductStatus) (*domain.OrderedProduct, error) {
	row := r.db.QueryRowContext(
		ctx,
		`UPDATE ordered_products 
		SET status = $1
		WHERE id = $2
		RETURNING id, product_id, session_id, status`,
		status,
		id,
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

func (r *OrderRepository) GetBillFromSession(ctx context.Context, id uuid.UUID) (*domain.Bill, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
    		p.id AS product_id,
    		p.name,
    		p.description, 
    		p.image_url,
    		p.delete_image_url,
    		p.category, 
    		p.price,
    		COUNT(op.id) as quantity,
    		(COUNT(op.id) * p.price) AS total_price
    	FROM ordered_products op
    	JOIN products p ON op.product_id = p.id
    	WHERE op.session_id = $1
    	GROUP BY p.id`,
		id,
	)
	defer func() {
		closeErr := rows.Close()
		if closeErr != nil {
			zap.L().Warn("error closing rows", zap.Error(closeErr))
		}
	}()

	if err != nil {
		zap.L().Error("error getting bill from session", zap.Error(err))
		return nil, domain.ErrInternal
	}

	var billItems []domain.BillItem
	var totalPrice decimal.Decimal
	for rows.Next() {
		var billItem domain.BillItem
		if err = rows.Scan(
			&billItem.Product.Id,
			&billItem.Product.Name,
			&billItem.Product.Description,
			&billItem.Product.ImageUrl,
			&billItem.Product.DeleteImageUrl,
			&billItem.Product.Category,
			&billItem.Product.Price,
			&billItem.Quantity,
			&billItem.TotalPrice,
		); err != nil {
			zap.L().Error("error scanning row", zap.Error(err))
			return nil, domain.ErrInternal
		}

		billItems = append(billItems, billItem)
		totalPrice = totalPrice.Add(billItem.TotalPrice)
	}

	return domain.NewBill(billItems, totalPrice), nil
}

func (r *OrderRepository) HasIncompletedOrderedProducts(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS(	
			SELECT id FROM ordered_products
			WHERE status != 'done' AND session_id = $1
    		LIMIT 1
    	)`,
		id,
	).Scan(&exists); err != nil {
		zap.L().Error("error scanning row", zap.Error(err))
		return false, domain.ErrInternal
	}

	return exists, nil
}
