package port

import (
	"context"
	"io"
	"restaurant/internal/core/domain"
)

type ImageRepository interface {
	// SaveImage saves an image and returns the URL to the image.
	SaveImage(ctx context.Context, data io.Reader) (*domain.Image, error)

	// DeleteImage deletes an image.
	DeleteImage(ctx context.Context, deleteUrl string) error
}
