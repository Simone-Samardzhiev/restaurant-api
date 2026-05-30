package storage

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"image"
	"image/png"
	"io"
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
		Bucket: aws.String(testS3BucketName),
		Key:    aws.String(key),
		Body:   strings.NewReader("fakeImage"),
	}); err != nil {
		t.Fatalf("Error uploading image: %v", err)
	}

	if err := storage.Delete(context.Background(), key); err != nil {
		t.Fatalf("Error deleting image: %v", err)
	}

	_, err := testS3Client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(testS3BucketName),
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
		Bucket: aws.String(testS3BucketName),
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

// generateTestImage generates a 10 * 10 png image.
func generateTestImage(t *testing.T) io.Reader {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	var buf bytes.Buffer

	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("Error encoding image: %v", err)
	}
	return &buf
}

func TestS3ImageStorageValidate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	storage := NewS3ImageStorage(testS3Client, 15*time.Second, testS3BucketName)

	imageKey := uuid.NewString()
	manager := transfermanager.New(testS3Client)
	if _, err := manager.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
		Bucket: aws.String(testS3BucketName),
		Key:    aws.String(imageKey),
		Body:   generateTestImage(t),
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
		t.Fatalf("Want error type domain.Error, got: %T", err)
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

func TestS3ImageStorageGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	storage := NewS3ImageStorage(testS3Client, 15*time.Second, testS3BucketName)
	imageKey := uuid.NewString()

	fakeImage := []byte("fakeImage")
	manager := transfermanager.New(testS3Client)
	if _, err := manager.UploadObject(context.Background(), &transfermanager.UploadObjectInput{
		Bucket: aws.String(testS3BucketName),
		Key:    aws.String(imageKey),
		Body:   bytes.NewReader(fakeImage),
	}); err != nil {
		t.Fatalf("Error uploading image: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		fetchedImage, err := storage.Get(context.Background(), imageKey)
		if err != nil {
			t.Fatalf("Error getting image: %v", err)
		}

		data, err := io.ReadAll(fetchedImage)
		if err != nil {
			t.Fatalf("Error reading image: %v", err)
		}
		if !bytes.Equal(data, fakeImage) {
			t.Fatalf("Want image data %s, got %s", fakeImage, data)
		}
	})
	t.Run("not found", func(t *testing.T) {
		_, err := storage.Get(context.Background(), uuid.NewString())
		if err == nil {
			t.Fatalf("Want error image not found, got nil")
		}

		if domainErr, ok := errors.AsType[*domain.Error](err); ok {
			if domainErr.Code != domain.ErrorCodeImageNotFound {
				t.Fatalf("Want error code %s, got %s", domain.ErrorCodeImageNotFound, domainErr.Code)
			}
			return
		}
		t.Fatalf("Want error type domain.Error, gor: %T", err)
	})
}
