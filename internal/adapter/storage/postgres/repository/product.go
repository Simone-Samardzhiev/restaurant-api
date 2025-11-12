package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"go.uber.org/zap"
)

// ProductRepository implements port.ProductRepository and provides access to postgres database.
type ProductRepository struct {
	db *sql.DB
}

// NewProductRepository creates a new ProductRepository instance.
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (r *ProductRepository) AddCategory(ctx context.Context, category *domain.ProductCategory) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO product_categories(id, name)
		VALUES ($1, $2)`,
		category.Id,
		category.Name,
	)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return domain.ErrProductCategoryNameAlreadyInUse
	} else if err != nil {
		zap.L().Error("error adding category", zap.Error(pqErr))
		return domain.ErrInternal
	}

	return nil
}

func (r *ProductRepository) UpdateCategory(ctx context.Context, dto *domain.UpdateCategoryProductDTO) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE product_categories
		SET name = COALESCE($1, name)
		WHERE id = $2`,
		dto.Name,
		dto.Id,
	)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return domain.ErrProductCategoryNameAlreadyInUse
	} else if err != nil {
		zap.L().Error("error updating category", zap.Error(pqErr))
		return domain.ErrInternal
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zap.L().Error("error getting rows affected", zap.Error(err))
		return domain.ErrInternal
	}

	if rows == 0 {
		return domain.ErrProductCategoryNotFound
	}
	return nil
}

func (r *ProductRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, "DELETE FROM product_categories WHERE id = $1", id)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23503" {
		return domain.ErrCategoryHasLinkedProducts
	} else if err != nil {
		zap.L().Error("error deleting category", zap.Error(pqErr))
		return domain.ErrInternal
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zap.L().Error("error getting rows affected", zap.Error(err))
		return domain.ErrInternal
	}

	if rows == 0 {
		return domain.ErrProductCategoryNotFound
	}
	return nil
}

func (r *ProductRepository) GetProductCategories(ctx context.Context) ([]domain.ProductCategory, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name FROM product_categories`)
	if err != nil {
		zap.L().Error("error getting product categories", zap.Error(err))
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			zap.L().Warn("error closing rows", zap.Error(closeErr))
		}
	}()

	var products []domain.ProductCategory

	for rows.Next() {
		var product domain.ProductCategory
		err = rows.Scan(&product.Id, &product.Name)
		if err != nil {
			zap.L().Error("error scanning rows", zap.Error(err))
			return nil, domain.ErrInternal
		}
		products = append(products, product)
	}

	return products, nil
}

var addProductPqErrorMap = map[string]map[string]error{
	"23505": {
		"products_name_key": domain.ErrProductNameAlreadyInUse,
	},
	"23503": {
		"products_category_fkey": domain.ErrProductCategoryNotFound,
	},
}

func (r *ProductRepository) AddProduct(ctx context.Context, product *domain.Product) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO 
    	products(id, name, description, image_url, delete_image_url ,category, price)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		product.Id,
		product.Name,
		product.Description,
		product.ImageUrl,
		product.DeleteImageUrl,
		product.Category,
		product.Price,
	)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if mappedCode, ok := addProductPqErrorMap[string(pqErr.Code)]; ok {
			if mappedConstraint, ok := mappedCode[pqErr.Constraint]; ok {
				return mappedConstraint
			}
		}

		zap.L().Error("unexpected pq error", zap.Error(pqErr))

	} else if err != nil {
		zap.L().Error(
			"error adding product",
			zap.String("id", product.Id.String()),
			zap.String("name", product.Name),
			zap.String("description", product.Description),
			zap.String("price", product.Price.String()),
			zap.String("categoryId", product.Category.String()),
			zap.Error(err),
		)
		return domain.ErrInternal
	}

	return nil
}

var updateProductPqErrorMap = map[string]map[string]error{
	"23505": {
		"products_name_key": domain.ErrProductNameAlreadyInUse,
	},
	"23503": {
		"products_category_fkey": domain.ErrProductCategoryNotFound,
	},
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, dto *domain.UpdateProductDTO) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE products
			SET name = COALESCE($1, name),
			description = COALESCE($2, description),
			category = COALESCE($3, category),
			price = COALESCE($4, price)
			WHERE id = $5`,
		dto.Name,
		dto.Description,
		dto.Category,
		dto.Price,
		dto.Id,
	)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if mappedCode, ok := updateProductPqErrorMap[string(pqErr.Code)]; ok {
			if mappedConstraint, ok := mappedCode[pqErr.Constraint]; ok {
				return mappedConstraint
			}
		}
		fmt.Println(pqErr.Constraint)
		fmt.Println(pqErr.Code)

		zap.L().Error("unexpected pq error", zap.Error(pqErr))
		return domain.ErrInternal
	} else if err != nil {
		zap.L().Error("error updating product", zap.Error(pqErr))
		return domain.ErrInternal
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zap.L().Error("error getting rows affected", zap.Error(err))
		return domain.ErrInternal
	}

	if rows == 0 {
		return domain.ErrProductCategoryNotFound
	}
	return nil
}

func (r *ProductRepository) UpdateProductImage(ctx context.Context, productId uuid.UUID, image *domain.Image) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE products
		SET image_url = $1,
		delete_image_url = $2
		WHERE id = $3`,
		image.Url,
		image.DeleteUrl,
		productId,
	)

	if err != nil {
		zap.L().Error("error updating product image", zap.Error(err))
		return domain.ErrInternal
	}

	rows, err := result.RowsAffected()
	if err != nil {
		zap.L().Error("error getting rows affected", zap.Error(err))
		return domain.ErrInternal
	}

	if rows == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func (r *ProductRepository) DeleteProductById(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	var sqlPath sql.NullString

	row := r.db.
		QueryRowContext(
			ctx,
			`DELETE FROM products 
       		WHERE id = $1 
       		RETURNING id, name, description, image_path, category, price`,
			id,
		)

	var product domain.Product
	err := row.Scan(&product.Id, &product.Name, &product.Description, &sqlPath, &product.Category, &product.Price)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	}
	if err != nil {
		zap.L().Error("error deleting product", zap.Error(err))
		return nil, domain.ErrInternal
	}

	if sqlPath.Valid {
		product.ImageUrl = &sqlPath.String
	} else {
		product.ImageUrl = nil
	}

	return &product, nil
}

func (r *ProductRepository) DeleteProductsByCategory(ctx context.Context, categoryId uuid.UUID) ([]domain.Product, error) {
	var products []domain.Product

	rows, err := r.db.QueryContext(
		ctx,
		`DELETE FROM products 
       	WHERE category = $1 
       	RETURNING id, name, description, image_path, category, price`,
		categoryId,
	)

	if err != nil {
		zap.L().Error("error deleting products", zap.Error(err))
		return nil, domain.ErrInternal
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			zap.L().Warn("error closing rows", zap.Error(closeErr))
		}
	}()

	for rows.Next() {
		var sqlPath sql.NullString
		var product domain.Product

		err = rows.Scan(&product.Id, &product.Name, &product.Description, &sqlPath, &product.Category, &product.Price)
		if err != nil {
			zap.L().Error("error scanning rows", zap.Error(err))
			return nil, domain.ErrInternal
		}

		if sqlPath.Valid {
			product.ImageUrl = &sqlPath.String
		} else {
			product.ImageUrl = nil
		}
		products = append(products, product)
	}
	return products, nil
}

func (r *ProductRepository) GetProductById(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT name, description, image_url, delete_image_url, category, price
		FROM products
		WHERE id = $1`,
		id,
	)

	var product domain.Product
	var imageUrl sql.NullString
	var deleteImageUrl sql.NullString

	err := row.Scan(
		&product.Name,
		&product.Description,
		&imageUrl,
		&deleteImageUrl,
		&product.Category,
		&product.Price,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrProductNotFound
	} else if err != nil {
		zap.L().Error("error getting product", zap.Error(err))
		return nil, domain.ErrInternal
	}

	if deleteImageUrl.Valid {
		product.ImageUrl = &deleteImageUrl.String
	} else {
		product.ImageUrl = nil
	}

	if deleteImageUrl.Valid {
		product.DeleteImageUrl = &deleteImageUrl.String
	} else {
		product.DeleteImageUrl = nil
	}

	product.Id = id
	return &product, nil
}

func (r *ProductRepository) GetProductsByCategory(ctx context.Context, categoryId uuid.UUID) ([]domain.Product, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, name, description, image_url, delete_image_url, price 
		FROM products
		WHERE category = $1`,
		categoryId,
	)
	if err != nil {
		zap.L().Error("error getting products", zap.Error(err))
		return nil, domain.ErrInternal
	}

	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			zap.L().Warn("error closing rows", zap.Error(closeErr))
		}
	}()

	var products []domain.Product
	var imageUrl sql.NullString
	var deleteImageUrl sql.NullString

	for rows.Next() {
		var product domain.Product
		err = rows.Scan(
			&product.Id,
			&product.Name,
			&product.Description,
			&imageUrl,
			&deleteImageUrl,
			&product.Price,
		)
		if err != nil {
			zap.L().Error("error scanning rows", zap.Error(err))
			return nil, domain.ErrInternal
		}

		product.Category = categoryId
		if imageUrl.Valid {
			product.ImageUrl = &imageUrl.String
		} else {
			product.ImageUrl = nil
		}

		if deleteImageUrl.Valid {
			product.ImageUrl = &deleteImageUrl.String
		} else {
			product.ImageUrl = nil
		}
		products = append(products, product)
	}
	return products, nil
}
