package port

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// ImageRepository is an interface for interacting with image data.
type ImageRepository interface {
	// Save saves an image with specified id and returns the path where ist saved.
	Save(ctx context.Context, image io.Reader, id uuid.UUID) (string, error)
}
