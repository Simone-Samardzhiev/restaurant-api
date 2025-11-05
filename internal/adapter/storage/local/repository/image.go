package repository

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"restaurant/internal/adapter/config"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ImageRepository implements port.ImageRepository and provides access to system local storage.
type ImageRepository struct {
	basePath string
}

// NewImageRepository creates a new ImageRepository instance.
func NewImageRepository(config *config.StorageConfig) *ImageRepository {
	return &ImageRepository{
		basePath: config.LocalFilesPath,
	}
}

func (r *ImageRepository) Save(_ context.Context, image io.Reader, id uuid.UUID) (string, error) {
	saveDir := filepath.Join(r.basePath, "images")
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		zap.L().Error(
			"error creating directory",
			zap.String("dir", saveDir),
			zap.Error(err),
		)
		return "", domain.ErrInternal
	}

	savePath := filepath.Join(saveDir, id.String()+".jpeg")
	file, err := os.Create(savePath)
	if err != nil {
		zap.L().Error(
			"error creating file",
			zap.String("file", savePath),
			zap.Error(err),
		)
		return "", domain.ErrInternal
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			zap.L().Error("error closing file", zap.Error(closeErr))
		}
	}()

	if _, err = io.Copy(file, image); err != nil {
		zap.L().Error("error copying image", zap.Error(err))
		return "", domain.ErrInternal
	}

	return savePath, nil
}
