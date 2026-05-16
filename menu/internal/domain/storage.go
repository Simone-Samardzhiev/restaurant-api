package domain

import "context"

// ImageStorage describes how image data is accessed.
type ImageStorage interface {

	// CreateUploadUrl creates an upload URL to which the client can upload the image.
	CreateUploadUrl(ctx context.Context, imageKey string, contentType ImageContentType) (string, error)
}
