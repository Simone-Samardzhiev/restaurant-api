package local_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"restaurant/internal/adapter/storage/local"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
	"testing"
)

func TestImageRepositoryCreateSavePath(t *testing.T) {
	tempDir := t.TempDir()
	repository := local.NewImageRepository(filepath.Join(tempDir, "image"))
	if err := repository.CreateSavePath(); err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}

func TestImageRepositorySaveImage(t *testing.T) {
	tempDir := t.TempDir()
	repository := local.NewImageRepository(tempDir)
	imageType, _ := menu.ParseImageType("jpg")

	if _, err := repository.SaveImage(context.Background(), bytes.NewBuffer([]byte{1, 2, 3, 34, 4, 5, 6}), imageType); err != nil {
		t.Fatalf("want no error, got %v", err)
	}
}

func TestImageRepositoryDeleteImage(t *testing.T) {
	tempDir := t.TempDir()
	repository := local.NewImageRepository(tempDir)
	imageType, _ := menu.ParseImageType("jpg")

	path, err := repository.SaveImage(context.Background(), bytes.NewBuffer([]byte{1, 2, 3, 34, 4, 5, 6}), imageType)
	if err != nil {
		t.Fatalf("want no error, got %v", err)
	}

	if err = repository.DeleteImage(context.Background(), path); err != nil {
		t.Fatalf("want no error, got %v", err)
	}

	if _, err = os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("want error %v, got %v", os.ErrNotExist, err)
		}
	}

	if err = repository.DeleteImage(context.Background(), path); err != nil {
		test.AssertError(t, err, domain.ErrorKindNotFound, domain.ErrorCodeImageNotFound)
	}
}
