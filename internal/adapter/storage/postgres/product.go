package postgres

import (
	"database/sql"
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"

	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

// ProductRepository implements [menu.ProductRepository] using postgres.
type ProductRepository struct {
	db *sql.DB
}

var _ menu.ProductRepository = (*ProductRepository)(nil)

// NewProductRepository allocates and creates a new [ProductRepository].
func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) SaveProduct(ctx context.Context, request *menu.SaveProductRequest) (*menu.Product, error) {
	product := &menu.Product{
		Id:          uuid.New(),
		Name:        request.Name,
		Description: request.Description,
		Price:       request.Price,
		CategoryId:  request.CategoryId,
		ImagePath:   request.ImagePath,
	}
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO products(id, name, description, category, price, image_path)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		product.Id,
		product.Name.String(),
		product.Description.String(),
		product.CategoryId.String(),
		product.Price.Value(),
		product.ImagePath,
	)

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "products_name_key" {
			return nil, domain.NewConflictError(
				"duplicate product name",
				domain.ErrorCodeProductNameConflict,
				err,
			)
		}

		if pqErr.Code == "23503" && pqErr.Constraint == "products_category_fkey" {
			return nil, &domain.Error{
				Kind:    domain.ErrorKindNotFound,
				Code:    domain.ErrorCodeCategoryNotFound,
				Message: "category not found",
				Details: []domain.ErrorDetail{
					{
						Code:    domain.ErrorCodeCategoryNotFoundByID,
						Message: "category not find by id",
						Metadata: map[string]any{
							"id": product.CategoryId,
						},
					},
				},
				Cause: err,
			}
		}
	}

	if err != nil {
		return nil, domain.NewInternalError("error inserting product", err)
	}

	return product, nil
}

func (r *ProductRepository) UpdateProduct(ctx context.Context, request *menu.UpdateProductRequest) error {
	var name sql.NullString
	var description sql.NullString
	var price decimal.NullDecimal
	var categoryId uuid.NullUUID
	var imagePath sql.NullString

	if request.Name != nil {
		name = sql.NullString{
			String: request.Name.String(),
			Valid:  true,
		}
	}
	if request.Description != nil {
		description = sql.NullString{
			String: request.Description.String(),
			Valid:  true,
		}
	}
	if request.Price != nil {
		price = decimal.NullDecimal{
			Decimal: request.Price.Value(),
			Valid:   true,
		}
	}
	if request.CategoryId != nil {
		categoryId = uuid.NullUUID{
			UUID:  *request.CategoryId,
			Valid: true,
		}
	}
	if request.ImagePath != nil {
		imagePath = sql.NullString{
			String: *request.ImagePath,
			Valid:  true,
		}
	}

	result, err := r.db.ExecContext(
		ctx,
		`UPDATE products 
		SET name = COALESCE($1, name),
		    description = COALESCE($2, description),
		    price = COALESCE($3, price),
		    category = COALESCE($4, category),
		    image_path = COALESCE($5, image_path)
		WHERE id = $6`,
		name,
		description,
		price,
		categoryId,
		imagePath,
		request.Id,
	)

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "products_name_key" {
			return domain.NewConflictError(
				"duplicate product name",
				domain.ErrorCodeProductNameConflict,
				err,
			)
		}

		if pqErr.Code == "23503" && pqErr.Constraint == "products_category_fkey" {
			return &domain.Error{
				Kind:    domain.ErrorKindNotFound,
				Code:    domain.ErrorCodeCategoryNotFound,
				Message: "category not found",
				Details: []domain.ErrorDetail{
					{
						Code:    domain.ErrorCodeCategoryNotFoundByID,
						Message: "category not find by id",
						Metadata: map[string]any{
							"id": request.CategoryId,
						},
					},
				},
				Cause: err,
			}
		}
	}

	if err != nil {
		return domain.NewInternalError("error updating product", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewInternalError("error getting rows affected", err)
	}
	if rows == 0 {
		return domain.NewNotFoundError(
			"product not found",
			domain.ErrorCodeProductNotFound,
			domain.ErrorDetail{
				Code:     domain.ErrorCodeProductNotFoundByID,
				Message:  "product not found by id",
				Metadata: map[string]interface{}{"id": request.Id},
			},
		)
	}

	return nil
}
