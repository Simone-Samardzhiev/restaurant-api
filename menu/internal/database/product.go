package database

import (
	"context"
	"database/sql"
	"errors"
	"menu/internal/domain"

	"github.com/lib/pq"
)

// PostgresProductRepository implements [domain.ProductRepository] using postgres.
type PostgresProductRepository struct {
	db *sql.DB
}

var _ domain.ProductRepository = (*PostgresProductRepository)(nil)

// NewPostgresProductRepository creates and allocates new [PostgresProductRepository].
func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{
		db: db,
	}
}

func (p *PostgresProductRepository) Save(ctx context.Context, product *domain.Product) error {
	_, err := p.db.ExecContext(
		ctx,
		`INSERT INTO products(id, name, description, price, category_id, image_key, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		product.Id,
		product.Name,
		product.Description,
		product.Price,
		product.CategoryId,
		product.ImageKey,
		product.Status,
		product.CreatedAt,
		product.UpdatedAt,
	)
	if err == nil {
		return nil
	}

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "products_name_key" {
			return domain.NewError("product name conflict", domain.ErrorCodeProductNameConflict, pqErr)
		}

		if pqErr.Code == "23503" && pqErr.Constraint == "products_category_id_fkey" {
			return domain.NewError("category not found", domain.ErrorCodeCategoryNotFound, pqErr)
		}
	}

	return domain.NewError("error saving product", domain.ErrorCodeInternal, err)
}
