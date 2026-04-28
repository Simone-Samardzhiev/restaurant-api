package db

import (
	"context"
	"database/sql"
	"errors"
	"menu/internal/domain"

	"github.com/lib/pq"
)

// CategoryRepository implements [domain.CategoryRepository]
// using postgres.
type CategoryRepository struct {
	db *sql.DB
}

var _ domain.CategoryRepository = (*CategoryRepository)(nil)

// NewCategoryRepository creates and allocates new [CategoryRepository].
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (c *CategoryRepository) Save(ctx context.Context, category *domain.Category) error {
	_, err := c.db.ExecContext(
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

func (c *CategoryRepository) GetAll(ctx context.Context) ([]domain.Category, error) {
	rows, err := c.db.QueryContext(ctx, "SELECT id, name, created_at, updated_at FROM categories")
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
