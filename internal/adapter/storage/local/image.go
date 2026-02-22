package local

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"

	"github.com/google/uuid"
)

// ImageRepository implements [menu.ImageRepository] using local file storage.
type ImageRepository struct {
	savePath string
}

var _ menu.ImageRepository = (*ImageRepository)(nil)

// NewImageRepository allocates and creates a new [ImageRepository].
func NewImageRepository(savePath string) *ImageRepository {
	return &ImageRepository{savePath: savePath}
}

func (r *ImageRepository) CreateSavePath() error {
	if err := os.MkdirAll(r.savePath, os.ModePerm); err != nil {
		return domain.NewInternalError("error creating save path", err)
	}
	return nil
}

func (r *ImageRepository) SaveImage(_ context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
	imagePath := uuid.NewString() + "." + imageType.String()
	file, err := os.Create(filepath.Join(r.savePath, imagePath))
	defer file.Close()
	if err != nil {
		return "", domain.NewInternalError("error creating image", err)
	}
	if _, err = io.Copy(file, data); err != nil {
		return "", domain.NewInternalError("error saving image", err)
	}
	return imagePath, nil
}

func (r *ImageRepository) DeleteImage(_ context.Context, path string) error {
	err := os.Remove(filepath.Join(r.savePath, path))
	if errors.Is(err, os.ErrNotExist) {
		return &domain.Error{
			Kind:    domain.ErrorKindNotFound,
			Code:    domain.ErrorCodeImageNotFound,
			Message: "image not found",
			Details: nil,
			Cause:   err,
		}
	} else if err != nil {
		return domain.NewInternalError("error deleting image", err)
	}
	return nil
}
