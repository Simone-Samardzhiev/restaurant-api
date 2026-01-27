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
	"github.com/shopspring/decimal"
)

// MenuRepository is the implementation of menu.Service using postgres/
type MenuRepository struct {
	db *sql.DB
}

var _ menu.Repository = (*MenuRepository)(nil)

// NewMenuRepository creates a new MenuRepository with database connection.
func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{
		db: db,
	}
}

func (r *MenuRepository) AddCategory(ctx context.Context, category *menu.Category) error {
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

func (r *MenuRepository) UpdateCategory(ctx context.Context, update *menu.CategoryUpdate) error {
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

func (r *MenuRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
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

func (r *MenuRepository) GetCategories(ctx context.Context, filter *menu.CategoryFilter) ([]menu.Category, error) {
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

func (r *MenuRepository) AddProduct(ctx context.Context, product *menu.Product) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO products (id, name, description, category, price, image_path) 
		VALUES ($1, $2, $3, $4, $5, $6)`,
		product.Id,
		product.Name.String(),
		product.Description.String(),
		product.CategoryId,
		product.Price.Value(),
		product.ImagePath,
	)

	if err == nil {
		return nil
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23503" {
			return domain.NewNotFoundError("category with id: " + product.CategoryId.String() + " not found")
		}

		if pqErr.Code == "23505" && pqErr.Constraint == "products_name_key" {
			return domain.NewConflictError("product with name: " + product.Name.String() + " already exists")
		}
	}

	return domain.NewInternalError(
		"error inserting product",
		err,
		domain.F("id", product.Id),
		domain.F("name", product.Name.String()),
		domain.F("description", product.Description.String()),
		domain.F("categoryId", product.CategoryId),
		domain.F("price", product.Price.Value()),
		domain.F("imagePath", product.ImagePath),
	)
}

func (r *MenuRepository) UpdateProduct(ctx context.Context, update *menu.ProductUpdate) error {
	var (
		name        sql.NullString
		description sql.NullString
		categoryId  sql.Null[uuid.UUID]
		price       sql.Null[decimal.Decimal]
	)

	if update.NewName != nil {
		name.String = update.NewName.String()
		name.Valid = true
	}

	if update.NewDescription != nil {
		description.String = update.NewDescription.String()
		description.Valid = true
	}

	if update.NewCategoryId != nil {
		categoryId.Valid = true
		categoryId.V = *update.NewCategoryId
	}

	if update.NewPrice != nil {
		price.Valid = true
		price.V = update.NewPrice.Value()
	}

	result, err := r.db.ExecContext(
		ctx,
		`UPDATE products
					SET name    = COALESCE($1, name),
    				description = COALESCE($2, description),
    				category    = COALESCE($3, category),
    				price       = COALESCE($4, price)
				WHERE id = $5`,
		name,
		description,
		categoryId,
		price,
		update.Id,
	)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23503" {
			return domain.NewNotFoundError("category with id: " + update.NewCategoryId.String() + " not found")
		}

		if pqErr.Code == "23505" && pqErr.Constraint == "products_name_key" {
			return domain.NewConflictError("product with name: " + update.NewName.String() + " already exists")
		}
	} else if err != nil {
		return domain.NewInternalError(
			"error updating product",
			err,
			domain.F("id", update.Id),
			domain.F("name", update.NewName),
			domain.F("description", update.NewDescription),
			domain.F("categoryId", update.NewCategoryId),
			domain.F("price", update.NewPrice),
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewInternalError("error getting rows affected", err)
	}

	if rows == 0 {
		return domain.NewNotFoundError("product with id: " + update.Id.String() + " not found")
	}

	return nil
}

func (r *MenuRepository) UpdateProductImagePath(ctx context.Context, id uuid.UUID, path string) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE products SET image_path = $1 WHERE id = $2`,
		path, id,
	)

	if err != nil {
		return domain.NewInternalError("error updating product image path", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewInternalError("error getting rows affected", err)
	}
	if rows == 0 {
		return domain.NewNotFoundError("product with id: " + id.String() + " not found")
	}

	return nil
}

func (r *MenuRepository) GetProductImagePaths(ctx context.Context) (map[string]struct{}, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT image_path FROM products`)
	if err != nil {
		return nil, domain.NewInternalError("error fetching product image paths", err)
	}
	defer rows.Close()

	imagePaths := make(map[string]struct{})
	for rows.Next() {
		var imagePath string
		if err = rows.Scan(&imagePath); err != nil {
			return nil, domain.NewInternalError("error scanning row", err)
		}
		imagePaths[imagePath] = struct{}{}
	}
	return imagePaths, nil
}

func (r *MenuRepository) GetProductImagePathById(ctx context.Context, id uuid.UUID) (string, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT image_path FROM products 
        WHERE id = $1`,
		id,
	)
	var path string

	err := row.Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.NewNotFoundError("product with id: " + id.String() + " not found")
	} else if err != nil {
		return "", domain.NewInternalError("error scanning row", err)
	}

	return path, nil
}

func (r *MenuRepository) GetProducts(ctx context.Context) ([]menu.Product, error) {
	var products []menu.Product

	rows, err := r.db.QueryContext(ctx, `SELECT id, name, description, category, price, image_path FROM products`)
	if err != nil {
		return nil, domain.NewInternalError("error fetching products", err)
	}

	defer rows.Close()

	for rows.Next() {
		var (
			id          uuid.UUID
			name        string
			description string
			categoryId  uuid.UUID
			price       decimal.Decimal
			imagePath   string
		)

		err = rows.Scan(&id, &name, &description, &categoryId, &price, &imagePath)
		if err != nil {
			return nil, domain.NewInternalError("error scanning row", err)
		}

		product, err := menu.NewProduct(id, name, description, categoryId, price, imagePath)
		if err != nil {
			return nil, domain.NewInternalError("error creating product", err)
		}
		products = append(products, *product)
	}

	return products, nil
}

func (r *MenuRepository) DeleteProduct(ctx context.Context, id uuid.UUID) (string, error) {
	row := r.db.QueryRowContext(
		ctx,
		`DELETE FROM products 
       WHERE id = $1
       RETURNING image_path`,
		id,
	)

	var path string
	err := row.Scan(&path)
	if errors.Is(err, sql.ErrNoRows) {
		return "", domain.NewNotFoundError("product with id: " + id.String() + " not found")
	} else if err != nil {
		return "", domain.NewInternalError("error scanning row", err)
	}

	return path, nil
}
