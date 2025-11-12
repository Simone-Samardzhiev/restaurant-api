package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"restaurant/internal/adapter/config"
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ImageRepository struct {
	apiKey string
}

func NewImageRepository(storageConfig *config.StorageConfig) *ImageRepository {
	return &ImageRepository{
		apiKey: storageConfig.ImagesApiKey,
	}
}

func (r *ImageRepository) createRequest(ctx context.Context, data io.Reader) (*http.Request, error) {
	baseUrl := "https://api.imgbb.com/1/upload"
	params := url.Values{}
	params.Set("key", r.apiKey)
	baseUrl += "?" + params.Encode()
	imageName := "restaurant-" + uuid.New().String()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", imageName)
	if err != nil {
		zap.L().Error(
			"error creating form file",
			zap.Error(err),
		)
		return nil, domain.ErrInternal
	}

	_, err = io.Copy(part, data)
	if err != nil {
		zap.L().Error(
			"error copping data",
			zap.Error(err),
		)
		return nil, domain.ErrInternal
	}

	if err = writer.Close(); err != nil {
		zap.L().Warn("error closing writer", zap.Error(err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseUrl, body)
	if err != nil {
		zap.L().Error(
			"error creating request",
			zap.Error(err),
		)
		return nil, domain.ErrInternal
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

type addImageResponse struct {
	Data struct {
		Url       string `json:"url"`
		DeleteUrl string `json:"delete_url"`
	} `json:"data"`
}

func (r *ImageRepository) SaveImage(ctx context.Context, data io.Reader) (*domain.Image, error) {
	req, err := r.createRequest(ctx, data)
	if err != nil {
		return nil, err
	}

	httpResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		zap.L().Error(
			"error executing request",
			zap.Error(err),
		)
	}
	defer func() {
		if closeErr := httpResponse.Body.Close(); closeErr != nil {
			zap.L().Warn(
				"error closing imageResponse body",
				zap.Error(closeErr),
			)
		}
	}()

	if httpResponse.StatusCode != http.StatusOK {
		zap.L().Error(
			"error executing request",
			zap.Int("status_code", httpResponse.StatusCode),
		)
		return nil, domain.ErrInternal
	}

	var imageResponse addImageResponse
	if err = json.NewDecoder(httpResponse.Body).Decode(&imageResponse); err != nil {
		zap.L().Error("error decoding imageResponse", zap.Error(err))
		return nil, domain.ErrInternal
	}

	return &domain.Image{
		Url:       imageResponse.Data.Url,
		DeleteUrl: imageResponse.Data.DeleteUrl,
	}, nil
}

func (r *ImageRepository) DeleteImage(ctx context.Context, deleteUrl string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, deleteUrl, nil)
	if err != nil {
		zap.L().Error("error creating request", zap.Error(err))
		return domain.ErrInternal
	}

	httpResponse, err := http.DefaultClient.Do(req)
	if err != nil {
		return domain.ErrInternal
	}
	if httpResponse.StatusCode != http.StatusOK {
		zap.L().Error("error executing request")
		return domain.ErrInternal
	}

	return nil
}
