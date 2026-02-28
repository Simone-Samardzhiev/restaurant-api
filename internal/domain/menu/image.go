package menu

import (
	"errors"
	"io"
	"restaurant/internal/domain"

	"github.com/google/uuid"
)

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

// UpdateImageRequest represents a request to update a product image.
type UpdateImageRequest struct {
	Id        uuid.UUID
	Data      io.Reader
	ImageType ImageType
}

// ParseUpdateImageRequest parses a [UpdateImageRequest] from id, image data and image type.
// If the fields are invalid the returned error will be of type [domain.Error].
func ParseUpdateImageRequest(id uuid.UUID, data io.Reader, imageType string) (*UpdateImageRequest, error) {
	parsedType, err := ParseImageType(imageType)
	if err != nil {
		if detailErr, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			return nil, domain.NewValidationError("invalid update image request", domain.ErrorCodeInvalidImageUpdate, *detailErr)
		}
		return nil, err
	}

	return &UpdateImageRequest{
		Id:        id,
		Data:      data,
		ImageType: parsedType,
	}, nil
}

// MustParseUpdateImageRequest is like [ParseUpdateImageRequest], but instead of
// returning an error it panics.
func MustParseUpdateImageRequest(id uuid.UUID, data io.Reader, imageType string) *UpdateImageRequest {
	request, err := ParseUpdateImageRequest(id, data, imageType)
	if err != nil {
		panic(err)
	}
	return request
}
