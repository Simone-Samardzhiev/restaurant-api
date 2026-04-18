package postgres

import (
	"context"
	"database/sql"
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// OrderedProductRepository implements [order.OrderedProductRepository] using postgres.
type OrderedProductRepository struct {
	db *sql.DB
}

var _ order.OrderedProductRepository = (*OrderedProductRepository)(nil)

// NewOrderedProductRepository allocates and creates a new [OrderedProductRepository].
func NewOrderedProductRepository(db *sql.DB) *OrderedProductRepository {
	return &OrderedProductRepository{
		db: db,
	}
}

func (r *OrderedProductRepository) Save(ctx context.Context, request *order.AddOrderedProductRequest) (*order.OrderedProduct, error) {
	product := &order.OrderedProduct{
		Id:        uuid.New(),
		ProductId: request.ProductId,
		SessionId: request.SessionId,
		Status:    order.StatusPending,
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO ordered_products (id, product_id, session_id, status)
		VALUES ($1, $2, $3, $4)`,
		product.Id,
		product.ProductId,
		product.SessionId,
		product.Status.String(),
	)

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23503" {
			if pqErr.Constraint == "ordered_products_product_id_fkey" {
				return nil, domain.NewNotFoundError("product not found", domain.ErrorCodeProductNotFound, domain.ErrorDetail{
					Code:     domain.ErrorCodeProductNotFoundByID,
					Message:  "product not found by id",
					Metadata: map[string]any{"product_id": product.Id},
				})
			}

			if pqErr.Constraint == "ordered_products_session_id_fkey" {
				return nil, domain.NewNotFoundError("session not found", domain.ErrorCodeSessionNotFound, domain.ErrorDetail{
					Code:     domain.ErrorCodeSessionNotFoundByID,
					Message:  "session not found",
					Metadata: map[string]any{"id": product.Id},
				})
			}
		}
	}

	if err != nil {
		return nil, domain.NewInternalError("error inserting ordered product", err)
	}

	return product, nil
}

func (r *OrderedProductRepository) GetBySessionId(ctx context.Context, sessionId uuid.UUID) ([]order.OrderedProduct, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, product_id, session_id, status
		FROM ordered_products
		WHERE session_id = $1`,
		sessionId,
	)

	if err != nil {
		return nil, domain.NewInternalError("error getting ordered products", err)
	}
	defer rows.Close()

	var products []order.OrderedProduct
	for rows.Next() {
		var id uuid.UUID
		var productId uuid.UUID
		var sessionId uuid.UUID
		var status string

		if err := rows.Scan(&id, &productId, &sessionId, &status); err != nil {
			return nil, domain.NewInternalError("error scanning row", err)
		}

		product, err := order.ParseOrderedProduct(id, productId, sessionId, status)
		if err != nil {
			return nil, domain.NewInternalError("error parsing ordered product", err)
		}

		products = append(products, *product)
	}

	return products, nil
}
