package postgres

import (
	"context"
	"database/sql"
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CategoryRepository implements [menu.CategoryRepository] using postgres.
type CategoryRepository struct {
	db *sql.DB
}

func (r *CategoryRepository) AddCategory(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error) {
	category := &menu.Category{
		Id:   uuid.New(),
		Name: request.Name,
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO product_categories (id, name) VALUES ($1, $2)`,
		category.Id, category.Name.String(),
	)

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "product_categories_name_key" {
			return nil, domain.NewConflictError(
				"duplicate category name",
				domain.ErrorCodeCategoryNameConflict,
				err,
			)
		}

	} else if err != nil {
		return nil, domain.NewInternalError("inserting category", err)
	}

	return category, nil
}

var _ menu.CategoryRepository = (*CategoryRepository)(nil)

func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}
