package storage

import (
	"context"
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
