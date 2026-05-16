package domain

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DefaultProductService is the default implementation of [ProductService].
type DefaultProductService struct {
	repository ProductRepository
	storage    ImageStorage
}

// NewProductService creates and allocates [DefaultProductService].
func NewProductService(repository ProductRepository, storage ImageStorage) *DefaultProductService {
	return &DefaultProductService{
		repository: repository,
		storage:    storage,
	}
}

var _ ProductService = (*DefaultProductService)(nil)

func (d *DefaultProductService) Add(ctx context.Context, request *AddProductRequest) (*ProductDraft, error) {
	now := time.Now()

	_, ext, ok := strings.Cut(string(request.ImageContentType), "/")
	if !ok {
		return nil, NewError("invalid image content type format: "+string(request.ImageContentType), ErrorCodeInternal, nil)
	}

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
