package local_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"restaurant/internal/adapter/storage/local"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"context"
)

func TestImageRepositoryAddImage(t *testing.T) {
	dir := t.TempDir()
	repository := local.NewImageRepository(dir)

	imageData := bytes.NewBuffer([]byte("Test data"))
	imageType, err := menu.NewImageType("png")
	if err != nil {
		t.Fatal("error creating image type", err)
	}

	path, err := repository.AddImage(context.Background(), imageData, imageType)
	if err != nil {
		t.Fatal("error adding image", err)
	}

	_, err = os.Stat(filepath.Join(dir, path))
	if err != nil {
		t.Fatal("image not saved correctly", err)
	}
}

func TestImageRepositoryDeleteImage(t *testing.T) {
	dir := t.TempDir()
	repository := local.NewImageRepository(dir)
	imageData := bytes.NewBuffer([]byte("Test data"))
	imageType, err := menu.NewImageType("png")
	if err != nil {
		t.Fatal("error creating image type", err)
	}

	path, err := repository.AddImage(context.Background(), imageData, imageType)
	if err != nil {
		t.Fatal("error adding image", err)
	}

	err = repository.DeleteImage(context.Background(), path)
	if err != nil {
		t.Fatal("error deleting image", err)
	}

	err = repository.DeleteImage(context.Background(), path)
	if err == nil {
		t.Fatal("expected error after deleting an deleted image", err)
	}

	var expectedErr *domain.Error
	if !errors.As(err, &expectedErr) {
		t.Fatal("expected error to be of type domain.Error")
	}

	if expectedErr.Type != domain.NotFound {
		t.Fatal("expected error to be of type NotFound")
	}
}

func TestImageRepositoryGetAllImagePaths(t *testing.T) {
	dir := t.TempDir()
	repository := local.NewImageRepository(dir)
	imageData := bytes.NewBuffer([]byte("Test data"))
	imageType, err := menu.NewImageType("png")
	if err != nil {
		t.Fatal("error creating image type", err)
	}

	paths := make(map[string]struct{}, 5)

	for i := 0; i < 5; i++ {
		path, err := repository.AddImage(context.Background(), imageData, imageType)
		if err != nil {
			t.Fatal("error adding image", err)
		}
		paths[path] = struct{}{}
	}

	result, err := repository.GetAllImagePaths(context.Background())
	if err != nil {
		t.Fatal("error getting all image paths", err)
	}

	if !reflect.DeepEqual(result, paths) {
		t.Fatal("expected results to be equal")
	}
}
