package database

import (
	"context"
	"database/sql"
	"errors"
	"menu/internal/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgresProductRepository implements [domain.ProductRepository] using postgres.
type PostgresProductRepository struct {
	db *sql.DB
}

func (p *PostgresProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := p.db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return domain.NewError("error deleting product", domain.ErrorCodeInternal, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewError("error getting rows affected", domain.ErrorCodeInternal, err)
	}
	if rows == 0 {
		return domain.NewError("product not found", domain.ErrorCodeProductNotFound, nil)
	}
	return nil
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
		`INSERT INTO products(id, name, description, price, category_id, image_key, image_content_type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		product.Id,
		product.Name,
		product.Description,
		product.Price,
		product.CategoryId,
		product.ImageKey,
		product.ImageContentType,
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

func (p *PostgresProductRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	row := p.db.QueryRowContext(ctx, `SELECT id, name, description, price, category_id, image_key, image_content_type, status, created_at, updated_at FROM products WHERE id = $1`, id)
	var product domain.Product
	err := row.Scan(&product.Id, &product.Name, &product.Description, &product.Price, &product.CategoryId, &product.ImageKey, &product.ImageContentType, &product.Status, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewError("product not found", domain.ErrorCodeProductNotFound, nil)
		}
		return nil, domain.NewError("error fetching product", domain.ErrorCodeInternal, err)
	}

	return &product, nil
}
