package domain

import (
	"context"
)

// ImageStorage describes how image data is accessed.
type ImageStorage interface {

	// CreateUploadUrl creates an upload URL to which the client can upload the image.
	CreateUploadUrl(ctx context.Context, imageKey string, contentType ImageContentType) (string, error)

	// Delete deletes an image by key.
	Delete(ctx context.Context, imageKey string) error

	// DeleteMultiple deletes multiple images by key.
	DeleteMultiple(ctx context.Context, imageKeys []string) error

	// Validate validates an image exists and the content type matches.
	Validate(ctx context.Context, imageKey string, contentType ImageContentType) error

	// Get fetches an image by key.
	Get(ctx context.Context, imageKey string) (*Image, error)
}
