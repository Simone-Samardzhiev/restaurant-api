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
		`INSERT INTO products_categories(id, name)
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
		`UPDATE products_categories
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
	result, err := r.db.ExecContext(ctx, "DELETE FROM products_categories WHERE id = $1", id)

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
    	products(id, name, description, image_path, category, price)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		product.Id,
		product.Name,
		product.Description,
		product.ImagePath,
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
			zap.String("description", product.Description),
			zap.String("imageId", product.ImagePath),
			zap.String("price", product.Price.String()),
			zap.String("categoryId", product.Category.String()),
			zap.String("name", product.Name),
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
