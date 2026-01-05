package postgres

import (
	"database/sql"
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/product"

	"github.com/lib/pq"
	"golang.org/x/net/context"
)

// ProductRepository is the implementation of product.Service using postgres/
type ProductRepository struct {
	db *sql.DB
}

var _ product.Repository = (*ProductRepository)(nil)

// NewProductRepository creates a new ProductRepository with database connection.
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) AddCategory(ctx context.Context, category *product.Category) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO product_categories (id, name) VALUES ($1, $2)",
		category.Id,
		category.RawName(),
	)

	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "product_categories_name_key" {
		return domain.NewConflictError("product category with this name already exists")
	}

	return domain.NewInternalError(
		"error inserting product category",
		err,
		domain.F("id", category.Id),
		domain.F("name", category.RawName()),
	)
}

func (r *ProductRepository) UpdateCategory(ctx context.Context, update *product.CategoryUpdate) error {
	var name sql.NullString
	if update.NewName != nil {
		name = sql.NullString{
			Valid:  true,
			String: update.NewName.Raw(),
		}
	}

	result, err := r.db.ExecContext(
		ctx,
		"UPDATE product_categories SET name = $1 WHERE id = $2",
		name,
		update.Id,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "product_categories_name_key" {
			return domain.NewConflictError("product category with this name already exists")
		}

		return domain.NewInternalError(
			"error updating product category",
			err,
			domain.F("id", update.Id),
			domain.F("name", update.NewName),
		)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.NewInternalError("error getting rows affected", err)
	}

	if rowsAffected == 0 {
		return domain.NewNotFoundError("product category not found")
	}
	return nil
}
