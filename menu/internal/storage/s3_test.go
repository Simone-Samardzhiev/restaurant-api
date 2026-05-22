package storage

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"menu/internal/domain"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/transfermanager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

func TestS3ImageStorageCreateUploadUrl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	storage := NewS3ImageStorage(testS3Client, 15*time.Second, testS3BucketName)

	url, err := storage.CreateUploadUrl(context.Background(), uuid.NewString()+".png", domain.ImageContentTypePNG)
	if err != nil {
		t.Fatalf("Error creating uploading URL: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader("fakeImage"))
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", domain.ImageContentTypePNG)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Error uploading image: %v", err)
	}
	res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("Unexpected status code: %d", res.StatusCode)
	}
}

func TestS3ImageStorageDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	storage := NewS3ImageStorage(testS3Client, 15*time.Second, testS3BucketName)

	key := uuid.NewString()
	manager := transfermanager.New(testS3Client)
	if _, err := manager.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
		Bucket: aws.String("images"),
		Key:    aws.String(key),
		Body:   strings.NewReader("fakeImage"),
	}); err != nil {
		t.Fatalf("Error uploading image: %v", err)
	}

	if err := storage.Delete(context.Background(), key); err != nil {
		t.Fatalf("Error deleting image: %v", err)
	}

	_, err := testS3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String("images"),
		Key:    aws.String(key),
	})
	if err == nil {
		t.Fatalf("Want not found error, got nil")
	}

	if _, ok := errors.AsType[*types.NotFound](err); !ok {
		t.Fatalf("Want *types.NotFound error, got: %T", err)
	}
}

func TestS3ImageStorageDeleteMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	storage := NewS3ImageStorage(testS3Client, 15*time.Second, testS3BucketName)

	key := uuid.NewString()
	manager := transfermanager.New(testS3Client)
	if _, err := manager.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
		Bucket: aws.String("images"),
		Key:    aws.String(key),
		Body:   strings.NewReader("fakeImage"),
	}); err != nil {
		t.Fatalf("Error uploading image: %v", err)
	}

	if err := storage.DeleteMultiple(context.Background(), []string{key}); err != nil {
		t.Fatalf("Error deleting images: %v", err)
	}

	_, err := testS3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String("images"),
		Key:    aws.String(key),
	})
	if err == nil {
		t.Fatalf("Want not found error, got nil")
	}

	if _, ok := errors.AsType[*types.NotFound](err); !ok {
		t.Fatalf("Want *types.NotFound error, got: %T", err)
	}
}

//go:embed testdata/french_fries.png
var image []byte

func TestS3ImageStorageValidate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	storage := NewS3ImageStorage(testS3Client, 15*time.Second, testS3BucketName)

	imageKey := uuid.NewString()
	manager := transfermanager.New(testS3Client)
	if _, err := manager.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
		Bucket: aws.String("images"),
		Key:    aws.String(imageKey),
		Body:   bytes.NewReader(image),
	}); err != nil {
		t.Fatalf("Error uploading image: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		if err := storage.Validate(context.Background(), imageKey, domain.ImageContentTypePNG); err != nil {
			t.Fatalf("Error validating image: %v", err)
		}
	})

	t.Run("invalid content type", func(t *testing.T) {
		err := storage.Validate(context.Background(), imageKey, domain.ImageContentTypeJPEG)
		if err == nil {
			t.Fatalf("Want error invalid image, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeInvalidImage {
				t.Fatalf("Want error code %s, got %s", domainErr.Code, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type domain.Error, gor: %T", err)
	})

	t.Run("not found", func(t *testing.T) {
		err := storage.Validate(context.Background(), uuid.NewString(), domain.ImageContentTypePNG)
		if err == nil {
			t.Fatalf("Want error image not found, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeImageNotFound {
				t.Fatalf("Want error code %s, got %s", domainErr.Code, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type domain.Error, gor: %T", err)
	})
}
