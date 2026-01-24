package local

import (
	"io"
	"os"
	"path/filepath"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"

	"context"

	"github.com/google/uuid"
)

type ImageRepository struct {
	savePath string
}

func NewImageRepository(savePath string) *ImageRepository {
	return &ImageRepository{
		savePath: savePath,
	}
}

var _ menu.ImageRepository = (*ImageRepository)(nil)

func (r *ImageRepository) CreateSavePath() error {
	return os.MkdirAll(r.savePath, os.ModePerm)
}

func (r *ImageRepository) AddImage(_ context.Context, data io.Reader, imageType menu.ImageType) (string, error) {
	imagePath := uuid.New().String() + "." + imageType.String()
	fullPath := filepath.Join(r.savePath, imagePath)

	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return "", domain.NewInternalError("error opening image file", err, domain.F("fullPath", fullPath))
	}

	defer file.Close()
	_, err = io.Copy(file, data)
	if err != nil {
		return "", domain.NewInternalError("error copying image file", err, domain.F("fullPath", fullPath))
	}

	return imagePath, nil
}

func (r *ImageRepository) DeleteImage(_ context.Context, path string) error {
	path = filepath.Join(r.savePath, path)
	err := os.Remove(path)

	if os.IsNotExist(err) {
		return domain.NewNotFoundError("image with path: " + path + " not found")
	}
	if err != nil {
		return domain.NewInternalError("error deleting image file", err, domain.F("path", path))
	}
	return nil
}

func (r *ImageRepository) GetAllImagePaths(_ context.Context) (map[string]struct{}, error) {
	result := make(map[string]struct{})

	content, err := os.ReadDir(r.savePath)
	if err != nil {
		return nil, domain.NewInternalError("error reading dir", err, domain.F("path", r.savePath))
	}

	for _, file := range content {
		if file.IsDir() {
			continue
		}
		result[file.Name()] = struct{}{}
	}
	return result, nil
}
