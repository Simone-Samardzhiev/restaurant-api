package postgres

import (
	"database/sql"
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"strings"

	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type CategoryRepository struct {
	db *sql.DB
}

var _ menu.CategoryRepository = (*CategoryRepository)(nil)

// NewCategoryRepository creates a new CategoryRepository.
func NewCategoryRepository(db *sql.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func (r *CategoryRepository) AddCategory(ctx context.Context, category *menu.Category) error {
	_, err := r.db.ExecContext(
		ctx,
		"INSERT INTO product_categories (id, name) VALUES ($1, $2)",
		category.Id,
		category.Name.String(),
	)

	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "product_categories_name_key" {
		return domain.NewConflictError("category with name: " + category.Name.String() + " already exists")
	}

	return domain.NewInternalError(
		"error inserting menu category",
		err,
		domain.F("id", category.Id),
		domain.F("name", category.Name.String()),
	)
}

func (r *CategoryRepository) UpdateCategory(ctx context.Context, update *menu.CategoryUpdate) error {
	var name sql.NullString
	if update.NewName != nil {
		name = sql.NullString{
			Valid:  true,
			String: update.NewName.String(),
		}
	}

	result, err := r.db.ExecContext(
		ctx,
		"UPDATE product_categories SET name = COALESCE($1, name) WHERE id = $2",
		name,
		update.Id,
	)

	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" && pqErr.Constraint == "product_categories_name_key" {
			return domain.NewConflictError("category with name: " + update.NewName.String() + " already exists")
		}

		return domain.NewInternalError(
			"error updating menu category",
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
		return domain.NewNotFoundError("category with id " + update.Id.String() + " not found")
	}
	return nil
}

func (r *CategoryRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM product_categories WHERE id = $1", id)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23503" && pqErr.Constraint == "products_category_fkey" {
			return domain.NewBadRequestError("category with id " + id.String() + " is used by products")
		}
	} else if err != nil {
		return domain.NewInternalError("error deleting menu category", err, domain.F("id", id))
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.NewInternalError("error getting rows affected", err)
	}

	if rowsAffected == 0 {
		return domain.NewNotFoundError("category with id: " + id.String() + " not found")
	}

	return nil
}

func (r *CategoryRepository) GetCategories(ctx context.Context, filter *menu.CategoryFilter) ([]menu.Category, error) {
	query := `SELECT id, name FROM product_categories`
	conditions := make([]string, 0)
	args := make([]interface{}, 0)

	if filter.Id != nil {
		conditions = append(conditions, "id = $1")
		args = append(args, *filter.Id)
	}

	if len(conditions) > 0 {
		query = query + " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, domain.NewInternalError("error fetching menu categories", err)
	}

	defer rows.Close()
	var categories []menu.Category
	for rows.Next() {
		var id uuid.UUID
		var name string

		if err = rows.Scan(&id, &name); err != nil {
			return nil, domain.NewInternalError("error scanning row", err)
		}

		category, err := menu.NewCategory(id, name)
		if err != nil {
			return nil, domain.NewInternalError("error creating category", err)
		}

		categories = append(categories, *category)
	}

	return categories, nil
}
