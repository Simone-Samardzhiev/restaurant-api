package postgres

import (
	"context"
	"database/sql"
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// CategoryRepository implements [menu.CategoryRepository] using postgres.
type CategoryRepository struct {
	db *sql.DB
}

var _ menu.CategoryRepository = (*CategoryRepository)(nil)

// NewCategoryRepository allocates and creates a new [CategoryRepository].
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) SaveCategory(ctx context.Context, request *menu.AddCategoryRequest) (*menu.Category, error) {
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

	}
	if err != nil {
		return nil, domain.NewInternalError("error inserting inserting category", err)
	}

	return category, nil
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, request *menu.UpdateCategoryRequest) error {
	var name sql.NullString
	if request.Name != nil {
		name.String = request.Name.String()
		name.Valid = true
	}

	result, err := r.db.ExecContext(
		ctx,
		`UPDATE product_categories SET name = COALESCE($1, name) WHERE id = $2`,
		name, request.Id,
	)

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "product_categories_name_key" {
			return domain.NewConflictError(
				"duplicate category name",
				domain.ErrorCodeCategoryNameConflict,
				err,
			)
		}
	}

	if err != nil {
		return domain.NewInternalError("error updating category", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewInternalError("error getting affected rows", err)
	}

	if rows == 0 {
		return domain.NewNotFoundError(
			"category not found",
			domain.ErrorCodeCategoryNotFound,
			domain.ErrorDetail{
				Code:     domain.ErrorCodeCategoryNotFoundByID,
				Message:  "category not found by id",
				Metadata: map[string]interface{}{"id": request.Id},
			},
		)
	}
	return nil
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(
		ctx,
		`DELETE FROM product_categories WHERE id = $1`,
		id,
	)

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23503" && pqErr.Constraint == "products_category_fkey" {
			return domain.NewConflictError(
				"cannot delete category with linked products",
				domain.ErrorCodeCategoryHasLinkedProducts,
				err,
			)
		}
	}

	if err != nil {
		return domain.NewInternalError("error deleting category", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewInternalError("error getting affected rows", err)
	}

	if rows == 0 {
		return domain.NewNotFoundError(
			"category not found",
			domain.ErrorCodeCategoryNotFound,
			domain.ErrorDetail{
				Code:     domain.ErrorCodeCategoryNotFoundByID,
				Message:  "category not found by id",
				Metadata: map[string]interface{}{"id": id},
			},
		)
	}
	return nil
}

func (r *CategoryRepository) GetCategories(ctx context.Context, filter *menu.CategoryFilter) ([]menu.Category, error) {
	var query strings.Builder
	var args []any
	query.WriteString("SELECT id, name FROM product_categories")

	if filter.Id != nil {
		query.WriteString(" WHERE id = $1")
		args = append(args, *filter.Id)
	}

	rows, err := r.db.QueryContext(ctx, query.String(), args...)
	defer rows.Close()

	if err != nil {
		return nil, domain.NewInternalError("error getting categories", err)
	}

	var categories []menu.Category
	for rows.Next() {
		var id uuid.UUID
		var name string

		if err = rows.Scan(&id, &name); err != nil {
			return nil, domain.NewInternalError("error scanning row", err)
		}

		category, err := menu.ParseCategory(id, name)
		if err != nil {
			return nil, domain.NewInternalError("error parsing category", err)
		}
		categories = append(categories, *category)
	}

	return categories, nil
}
