package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"menu/internal/domain"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3ImageStorage implements [domain.ImageStorage] using S3 bucket.
type S3ImageStorage struct {
	client        *s3.Client
	presignClient *s3.PresignClient

	urlExpiry time.Duration
	bucket    string
}

var _ domain.ImageStorage = (*S3ImageStorage)(nil)

// NewS3ImageStorage creates and allocates new [S3ImageStorage].
func NewS3ImageStorage(client *s3.Client, urlExpiryTine time.Duration, bucket string) *S3ImageStorage {
	return &S3ImageStorage{
		client:        client,
		presignClient: s3.NewPresignClient(client),
		urlExpiry:     urlExpiryTine,
		bucket:        bucket,
	}
}

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

func (s *S3ImageStorage) DeleteMultiple(ctx context.Context, imageKeys []string) error {
	objects := make([]types.ObjectIdentifier, 0, len(imageKeys))
	for _, imageKey := range imageKeys {
		objects = append(objects, types.ObjectIdentifier{
			Key: aws.String(imageKey),
		})
	}

	var errs []error
	for i := 0; i < len(objects); i += 1000 {
		end := i + 1000
		if end > len(objects) {
			end = len(objects)
		}
		batch := objects[i:end]

		out, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(s.bucket),
			Delete: &types.Delete{
				Objects: batch,
			},
		})
		if err != nil {
			errs = append(errs, fmt.Errorf("batch deletion failed at index %d: %w", i, err))
			continue
		}

		if len(out.Errors) > 0 {
			for _, err := range out.Errors {
				errs = append(errs, fmt.Errorf("error deleting image with key %s: %s", *err.Key, *err.Message))
			}
		}
	}

	if len(errs) > 0 {
		return domain.NewError("error deleting images", domain.ErrorCodeInternal, errors.Join(errs...))
	}
	return nil
}

func (s *S3ImageStorage) Validate(ctx context.Context, imageKey string, contentType domain.ImageContentType) error {
	res, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(imageKey),
		Range:  aws.String("bytes=0-511"),
	})
	if err != nil {
		if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
			return domain.NewError("image not found", domain.ErrorCodeImageNotFound, nil)
		}
		if _, ok := errors.AsType[*types.NotFound](err); ok {
			return domain.NewError("image not found", domain.ErrorCodeImageNotFound, nil)
		}

		return domain.NewError("error getting image", domain.ErrorCodeInternal, err)
	}
	defer res.Body.Close()

	buffer, err := io.ReadAll(res.Body)
	if err != nil {
		return domain.NewError("error reading image", domain.ErrorCodeInternal, err)
	}

	content := http.DetectContentType(buffer)
	if content != string(contentType) {
		return domain.NewError("invalid image content type", domain.ErrorCodeInvalidImage, nil)
	}

	return nil
}

func (s *S3ImageStorage) Get(ctx context.Context, imageKey string) (*domain.Image, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(imageKey),
	})
	if err == nil {
		var contentType domain.ImageContentType = domain.ImageContentTypeJPEG
		if out.ContentType != nil {
			contentType = domain.ImageContentType(*out.ContentType)
		}

		return &domain.Image{
			Data:        out.Body,
			ContentType: contentType,
		}, nil
	}

	if _, ok := errors.AsType[*types.NoSuchKey](err); ok {
		return nil, domain.NewError("image not found", domain.ErrorCodeImageNotFound, nil)
	}
	if _, ok := errors.AsType[*types.NotFound](err); ok {
		return nil, domain.NewError("image not found", domain.ErrorCodeImageNotFound, nil)
	}

	return nil, domain.NewError("error getting image", domain.ErrorCodeInternal, err)
}
