package domain

import "context"

// ImageStorage describes how image data is accessed.
type ImageStorage interface {

	// CreateUploadUrl creates an upload URL to which the client can uplaod the image.
	CreateUploadUrl(ctx context.Context, imageKey, contentType string) (string, error)
}
