package storage

import (
	"context"
	"menu/internal/domain"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3ImageStorage implements [domain.ImageStorage] using S3 bucket.
type S3ImageStorage struct {
	client        *s3.Client
	presignClient *s3.PresignClient

	urlExpiry time.Duration
	bucket    string
}

// NewS3ImageStorage creates and allocates new [S3ImageStorage].
func NewS3ImageStorage(client *s3.Client, urlExpiryTine time.Duration, bucket string) *S3ImageStorage {
	return &S3ImageStorage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		urlExpiry:     urlExpiryTine,
		bucket:        bucket,
	}
}

var _ domain.ImageStorage = (*S3ImageStorage)(nil)

func (s *S3ImageStorage) CreateUploadUrl(ctx context.Context, imageKey string, contentType domain.ImageContentType) (string, error) {
	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(imageKey),
		ContentType: aws.String(string(contentType)),
	}, func(options *s3.PresignOptions) {
		options.Expires = s.urlExpiry
	})

	if err != nil {
		return "", domain.NewError("error presigning url for image upload", domain.ErrorCodeInternal, err)
	}

	return req.URL, nil
}

func (s *S3ImageStorage) Delete(ctx context.Context, imageKey string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(imageKey),
	})

	if err == nil {
		return nil
	}
	return domain.NewError("error deleting image", domain.ErrorCodeInternal, err)
}
