package repository

import (
	"context"
	"database/sql"
	"errors"
	"restaurant/internal/core/domain"

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
		errorMap := map[string]map[string]error{
			"23505": {
				"products_name_key": domain.ErrProductNameAlreadyInUse,
			},
			"23503": {
				"products_category_fkey": domain.ErrProductCategoryNotFound,
			},
		}

		if mappedCode, ok := errorMap[pqErr.Code.Name()]; ok {
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
