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
