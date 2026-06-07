package database

import (
	"context"
	"database/sql"
	"errors"
	"menu/internal/domain"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// PostgresProductRepository implements [domain.ProductRepository] using postgres.
type PostgresProductRepository struct {
	db *sql.DB
}

var _ domain.ProductRepository = (*PostgresProductRepository)(nil)

// NewPostgresProductRepository creates and allocates new [PostgresProductRepository].
func NewPostgresProductRepository(db *sql.DB) *PostgresProductRepository {
	return &PostgresProductRepository{
		db: db,
	}
}

func (p *PostgresProductRepository) Save(ctx context.Context, product *domain.Product) error {
	_, err := p.db.ExecContext(
		ctx,
		`INSERT INTO products(id, name, description, price, category_id, image_key, image_content_type, status, pending_image_key, pending_image_content_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		product.Id,
		product.Name,
		product.Description,
		product.Price,
		product.CategoryId,
		product.ImageKey,
		product.ImageContentType,
		product.Status,
		product.PendingImageKey,
		product.PendingImageContentType,
		product.CreatedAt,
		product.UpdatedAt,
	)
	if err == nil {
		return nil
	}

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "products_name_key" {
			return domain.NewError("product name conflict", domain.ErrorCodeProductNameConflict, pqErr)
		}

		if pqErr.Code == "23503" && pqErr.Constraint == "products_category_id_fkey" {
			return domain.NewError("category not found", domain.ErrorCodeCategoryNotFound, pqErr)
		}
	}

	return domain.NewError("error saving product", domain.ErrorCodeInternal, err)
}

func (p *PostgresProductRepository) Get(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	row := p.db.QueryRowContext(ctx, `SELECT id, name, description, price, category_id, image_key, image_content_type, status, pending_image_key, pending_image_content_type, created_at, updated_at FROM products WHERE id = $1`, id)
	var product domain.Product
	var pendingImageKey sql.NullString
	var pendingImageContentType sql.NullString

	err := row.Scan(&product.Id, &product.Name, &product.Description, &product.Price, &product.CategoryId, &product.ImageKey, &product.ImageContentType, &product.Status, &pendingImageKey, &pendingImageContentType, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewError("product not found", domain.ErrorCodeProductNotFound, nil)
		}
		return nil, domain.NewError("error fetching product", domain.ErrorCodeInternal, err)
	}

	if pendingImageKey.Valid {
		product.PendingImageKey = &pendingImageKey.String
	}
	if pendingImageContentType.Valid {
		product.PendingImageContentType = new(domain.ImageContentType(pendingImageContentType.String))
	}

	return &product, nil
}

func (p *PostgresProductRepository) GetAllWithImage(ctx context.Context) ([]domain.Product, error) {
	rows, err := p.db.QueryContext(ctx, `SELECT id, name, description, price, category_id, image_key, image_content_type, status, pending_image_key, pending_image_content_type, created_at, updated_at FROM products WHERE status != 'missing_image'`)
	if err != nil {
		return nil, domain.NewError("error fetching all ready products", domain.ErrorCodeInternal, err)
	}

	defer rows.Close()
	var products []domain.Product
	for rows.Next() {
		var product domain.Product
		var pendingImageKey sql.NullString
		var pendingImageContentType sql.NullString

		if err := rows.Scan(&product.Id, &product.Name, &product.Description, &product.Price, &product.CategoryId, &product.ImageKey, &product.ImageContentType, &product.Status, &pendingImageKey, &pendingImageContentType, &product.CreatedAt, &product.UpdatedAt); err != nil {
			return nil, domain.NewError("error scanning product", domain.ErrorCodeInternal, err)
		}
		if pendingImageKey.Valid {
			product.PendingImageKey = &pendingImageKey.String
		}
		if pendingImageContentType.Valid {
			product.PendingImageContentType = new(domain.ImageContentType(pendingImageContentType.String))
		}

		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, domain.NewError("error scanning products", domain.ErrorCodeInternal, err)
	}

	return products, nil
}

func (p *PostgresProductRepository) Update(ctx context.Context, request *domain.UpdateProductRequest) error {
	updates := make([]string, 0, 4)
	args := make([]any, 0, 5)

	if request.Name != nil {
		updates = append(updates, "name = $"+strconv.Itoa(len(updates)+1))
		args = append(args, *request.Name)
	}
	if request.Description != nil {
		updates = append(updates, "description = $"+strconv.Itoa(len(updates)+1))
		args = append(args, *request.Description)
	}
	if request.Price != nil {
		updates = append(updates, "price = $"+strconv.Itoa(len(updates)+1))
		args = append(args, *request.Price)
	}
	if request.CategoryId != nil {
		updates = append(updates, "category_id = $"+strconv.Itoa(len(updates)+1))
		args = append(args, *request.CategoryId)
	}

	if len(updates) == 0 {
		return nil
	}

	query := "UPDATE products SET " + strings.Join(updates, ", ") + " WHERE id = $" + strconv.Itoa(len(updates)+1)
	args = append(args, request.Id)

	result, err := p.db.ExecContext(ctx, query, args...)
	if err == nil {
		rows, err := result.RowsAffected()
		if err != nil {
			return domain.NewError("error getting rows affected", domain.ErrorCodeInternal, err)
		}
		if rows == 0 {
			return domain.NewError("product not found", domain.ErrorCodeProductNotFound, nil)
		}

		return nil
	}

	if pqErr, ok := errors.AsType[*pq.Error](err); ok {
		if pqErr.Code == "23505" && pqErr.Constraint == "products_name_key" {
			return domain.NewError("product name conflict", domain.ErrorCodeProductNameConflict, pqErr)
		}

		if pqErr.Code == "23503" && pqErr.Constraint == "products_category_id_fkey" {
			return domain.NewError("category not found", domain.ErrorCodeCategoryNotFound, pqErr)
		}
	}

	return domain.NewError("error updating product", domain.ErrorCodeInternal, err)
}

func (p *PostgresProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := p.db.ExecContext(ctx, "DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return domain.NewError("error deleting product", domain.ErrorCodeInternal, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewError("error getting rows affected", domain.ErrorCodeInternal, err)
	}
	if rows == 0 {
		return domain.NewError("product not found", domain.ErrorCodeProductNotFound, nil)
	}
	return nil
}

func (p *PostgresProductRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ProductStatus) error {
	result, err := p.db.ExecContext(ctx, "UPDATE products SET status = $1 WHERE id = $2", status, id)
	if err != nil {
		return domain.NewError("error updating product", domain.ErrorCodeInternal, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return domain.NewError("error getting rows affected", domain.ErrorCodeInternal, err)
	}

	if rows == 0 {
		return domain.NewError("product not found", domain.ErrorCodeProductNotFound, nil)
	}
	return nil
}

func (p *PostgresProductRepository) DeleteExpiredByStatus(ctx context.Context, olderThan time.Duration) ([]string, error) {
	t := time.Now().Add(-olderThan)
	rows, err := p.db.QueryContext(
		ctx,
		`DELETE FROM products 
    	WHERE status != 'ready' AND created_at < $1 
    	RETURNING image_key`, t,
	)
	if err != nil {
		return nil, domain.NewError("error deleting products", domain.ErrorCodeInternal, err)
	}
	defer rows.Close()

	var imageKeys []string
	for rows.Next() {
		var imageKey string
		if err := rows.Scan(&imageKey); err != nil {
			return nil, domain.NewError("error scanning row", domain.ErrorCodeInternal, err)
		}

		imageKeys = append(imageKeys, imageKey)
	}
	if err = rows.Err(); err != nil {
		return nil, domain.NewError("error scanning rows", domain.ErrorCodeInternal, err)
	}
	return imageKeys, nil
}
