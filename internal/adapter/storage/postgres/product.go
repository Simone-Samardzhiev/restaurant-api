package postgres

import (
	"database/sql"
	"errors"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"

	"context"

	"github.com/google/uuid"
	"github.com/lib/pq"
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

func (r ProductRepository) SaveProduct(ctx context.Context, request *menu.SaveProductRequest) (*menu.Product, error) {
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
