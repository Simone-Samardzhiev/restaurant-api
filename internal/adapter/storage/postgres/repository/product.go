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
	var imagePath sql.NullString
	if product.ImagePath != nil {
		imagePath.Valid = true
		imagePath.String = *product.ImagePath
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO 
    	products(id, name, description, image_path, category, price)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		product.Id,
		product.Name,
		product.Description,
		imagePath.String,
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
			zap.Any("imagePath", imagePath.String),
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

func (r *ProductRepository) UpdateProductImagePath(ctx context.Context, id uuid.UUID, path *string) error {
	var sqlPath sql.NullString
	if path != nil {
		sqlPath.Valid = true
		sqlPath.String = *path
	}

	var result bool
	err := r.db.QueryRowContext(
		ctx,
		`UPDATE products SET image_path = $1 
        WHERE id = $2
        RETURNING TRUE`,
		sqlPath,
		id,
	).Scan(&result)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrProductNotFound
	} else if err != nil {
		zap.L().Error("error updating product image path", zap.Error(err))
		return domain.ErrInternal
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
		product.ImagePath = &sqlPath.String
	} else {
		product.ImagePath = nil
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
			product.ImagePath = &sqlPath.String
		} else {
			product.ImagePath = nil
		}
		products = append(products, product)
	}
	return products, nil
}
