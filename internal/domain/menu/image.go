package menu

import (
	"fmt"
)

// ImageType represents a valid image type.
type ImageType struct {
	raw string
}

func (i *ImageType) String() string {
	return i.raw
}

var imageTypes = map[string]struct{}{
	"jpeg": {},
	"jpg":  {},
	"png":  {},
	"webp": {},
}

// NewImageType parses a ImageType from string.
func NewImageType(imageType string) (ImageType, error) {
	if _, ok := imageTypes[imageType]; ok {
		return ImageType{imageType}, nil
	}

	return ImageType{}, fmt.Errorf("invalid image type: %s", imageType)
}
