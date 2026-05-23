package domain

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DefaultProductService is the default implementation of [ProductService].
type DefaultProductService struct {
	repository ProductRepository
	storage    ImageStorage
	logger     *slog.Logger
}

// NewDefaultProductService creates and allocates [DefaultProductService].
func NewDefaultProductService(repository ProductRepository, storage ImageStorage, logger *slog.Logger) *DefaultProductService {
	return &DefaultProductService{
		repository: repository,
		storage:    storage,
		logger:     logger,
	}
}

var _ ProductService = (*DefaultProductService)(nil)

func (d *DefaultProductService) Add(ctx context.Context, request *AddProductRequest) (*ProductDraft, error) {
	_, ext, ok := strings.Cut(string(request.ImageContentType), "/")
	if !ok {
		return nil, NewError("invalid image content type format: "+string(request.ImageContentType), ErrorCodeInternal, nil)
	}

	now := time.Now()
	productId := uuid.New()
	product := Product{
		Id:               productId,
		Name:             request.Name,
		Description:      request.Description,
		Price:            request.Price,
		CategoryId:       request.CategoryId,
		ImageKey:         productId.String() + "." + ext,
		ImageContentType: request.ImageContentType,
		Status:           ProductStatusAwaitingImage,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := d.repository.Save(ctx, &product); err != nil {
		return nil, err
	}

	uploadUrl, err := d.storage.CreateUploadUrl(ctx, product.ImageKey, product.ImageContentType)
	if err != nil {
		return nil, err
	}

	return &ProductDraft{
		Id:             productId,
		ImageUploadUrl: uploadUrl,
	}, nil
}

// cleanUpProduct deletes the product from the repository and the image from the storage.
// Any errors during the process are logged.
func (d *DefaultProductService) cleanUpProduct(ctx context.Context, product *Product) {
	if deleteErr := d.repository.Delete(ctx, product.Id); deleteErr != nil {
		d.logger.LogAttrs(
			ctx, slog.LevelWarn,
			"Error deleting product record for cleanup",
			slog.String("error", deleteErr.Error()),
		)
	}

	if deleteErr := d.storage.Delete(ctx, product.ImageKey); deleteErr != nil {
		d.logger.LogAttrs(
			ctx, slog.LevelWarn,
			"Error deleting product image for cleanup",
			slog.String("error", deleteErr.Error()),
		)
	}
}

func (d *DefaultProductService) ConfirmImageUpload(ctx context.Context, productID uuid.UUID) error {
	product, err := d.repository.Get(ctx, productID)
	if err != nil {
		return err
	}

	// If the product image has already been confirmed return nil.
	if product.Status == ProductStatusReady {
		return nil
	}

	if err = d.storage.Validate(ctx, product.ImageKey, product.ImageContentType); err != nil {
		domainErr, ok := errors.AsType[*Error](err)
		if ok {
			if domainErr.Code == ErrorCodeInvalidImage {
				d.cleanUpProduct(ctx, product)
			}
		}
		return err
	}

	if err = d.repository.UpdateStatus(ctx, product.Id, product.Status); err != nil {
		if deleteErr := d.storage.Delete(ctx, product.ImageKey); deleteErr != nil {
			slog.Default().LogAttrs(
				ctx, slog.LevelWarn,
				"Image could not be deleted after failing to update product status",
				slog.String("error", deleteErr.Error()),
			)
		}
		return err
	}

	return nil
}
