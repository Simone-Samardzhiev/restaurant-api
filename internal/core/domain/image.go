package domain

import "io"

// ImageType is an enum for different images types.
type ImageType string

// Image represents an image entity.
type Image struct {
	Data io.Reader
	Type ImageType
}

// NewImage creates a new Image instace.
func NewImage(data io.Reader, imageType ImageType) *Image {
	return &Image{
		Data: data,
		Type: imageType,
	}
}
