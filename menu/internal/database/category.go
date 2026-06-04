package database

import (
	"context"
	"database/sql"
	"errors"
	"menu/internal/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgresCategoryRepository implements [domain.CategoryRepository]
// using postgres.
type PostgresCategoryRepository struct {
	db *sql.DB
}

var _ domain.CategoryRepository = (*PostgresCategoryRepository)(nil)

// NewPostgresCategoryRepository creates and allocates new [PostgresCategoryRepository].
func NewPostgresCategoryRepository(db *sql.DB) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{db: db}
}

func (p *PostgresCategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	_, err := p.db.ExecContext(
		ctx,
		`INSERT INTO categories (id, name, created_at, updated_at) 
		VALUES ($1, $2, $3, $4)`,
		category.Id,
		category.Name,
		category.CreatedAt,
		category.UpdatedAt,
	)

	if err == nil {
		return nil
	}

	pqErr, ok := errors.AsType[*pq.Error](err)
	if ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "categories_name_key" {
			return domain.NewError("category name conflict", domain.ErrorCodeCategoryNameConflict, err)
		}
	}

	return domain.NewError("error saving category", domain.ErrorCodeInternal, err)
}

func (p *PostgresCategoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT id, name, created_at, updated_at FROM categories")
	if err != nil {
		return nil, domain.NewError("error getting categories", domain.ErrorCodeInternal, err)
	}
	defer rows.Close()

	var categories []domain.Category
	for rows.Next() {
		var category domain.Category
		if err := rows.Scan(&category.Id, &category.Name, &category.CreatedAt, &category.UpdatedAt); err != nil {
			return nil, domain.NewError("error scanning row", domain.ErrorCodeInternal, err)
		}

		categories = append(categories, category)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.NewError("error scanning rows", domain.ErrorCodeInternal, err)
	}
	return categories, nil
}

func (p *PostgresCategoryRepository) Update(ctx context.Context, id uuid.UUID, name string) error {
	result, err := p.db.ExecContext(ctx, "UPDATE categories SET name = $1 WHERE id = $2", name, id)
	if err == nil {
		rows, err := result.RowsAffected()
		if err != nil {
			return domain.NewError("error getting rows affected", domain.ErrorCodeInternal, err)
		}

		if rows == 0 {
			return domain.NewError("category not found", domain.ErrorCodeCategoryNotFound, nil)
		}

		return nil
	}

	pqErr, ok := errors.AsType[*pq.Error](err)
	if ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "categories_name_key" {
			return domain.NewError("category name conflict", domain.ErrorCodeCategoryNameConflict, err)
		}
	}

	return domain.NewError("error updating category", domain.ErrorCodeInternal, err)
}

func (p *PostgresCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := p.db.ExecContext(ctx, "DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		if pqErr, ok := errors.AsType[*pq.Error](err); ok {
			if pqErr.Code == "23503" && pqErr.Constraint == "products_category_id_fkey" {
				return domain.NewError("category has linked products", domain.ErrorCodeCategoryHasProducts, err)
			}
		}
		return domain.NewError("error deleting category", domain.ErrorCodeInternal, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewError("error getting rows affected", domain.ErrorCodeInternal, err)
	}

	if rows == 0 {
		return domain.NewError("category not found", domain.ErrorCodeCategoryNotFound, nil)
	}
	return nil
}
