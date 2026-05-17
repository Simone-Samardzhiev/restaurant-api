package storage

import (
	"context"
	"menu/internal/domain"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)
import "github.com/aws/aws-sdk-go-v2/config"

func TestS3ImageStorageCreateUploadUrl(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatalf("Error loading aws config: %v", err)
	}

	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	storage := NewS3ImageStorage(client, 15*time.Second, "images")

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
