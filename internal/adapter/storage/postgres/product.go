package postgres

import (
	"context"
	"database/sql"
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

type ProductRepository struct {
	db *sql.DB
}

var _ menu.ProductRepository = (*ProductRepository)(nil)

// NewProductRepository creates a new ProductRepository.
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) AddProduct(ctx context.Context, product *menu.Product) error {
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

func (r *ProductRepository) UpdateProduct(ctx context.Context, update *menu.ProductUpdate) error {
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

func (r *ProductRepository) UpdateProductImagePath(ctx context.Context, id uuid.UUID, path string) error {
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

func (r *ProductRepository) GetProductImagePathById(ctx context.Context, id uuid.UUID) (string, error) {
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

func (r *ProductRepository) GetProductImagePaths(ctx context.Context) (map[string]struct{}, error) {
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

func (r *ProductRepository) GetProducts(ctx context.Context) ([]menu.Product, error) {
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

func (r *ProductRepository) DeleteProduct(ctx context.Context, id uuid.UUID) (string, error) {
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
