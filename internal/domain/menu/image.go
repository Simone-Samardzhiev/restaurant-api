package menu

import "restaurant/internal/domain"

// ImageType represents a supported image format.
type ImageType struct {
	raw string
}

// ParseImageType parses an [ImageType] from string.
//
// If the image type is invalid the returned error will be of type [domain.ErrorDetail].
func ParseImageType(imageType string) (ImageType, error) {
	switch imageType {
	case "png", "jpg", "jpeg":
		return ImageType{imageType}, nil
	default:
		return ImageType{}, &domain.ErrorDetail{
			Code:    domain.ErrorCodeInvalidImageType,
			Message: "invalid image type",
			Metadata: map[string]any{
				"actual":    imageType,
				"supported": []string{"png", "jpg", "jpeg"},
			},
		}
	}
}

func (i *ImageType) String() string {
	return i.raw
}
