package port

import (
	"context"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
)

// ImageRepository is an interface for interacting with image data.
type ImageRepository interface {
	// Save saves an image with specified id and returns the path where ist saved.
	Save(ctx context.Context, image *domain.Image, id uuid.UUID) (string, error)
	// Delete deletes an image with specified id.
	Delete(ctx context.Context, path string) error
}
