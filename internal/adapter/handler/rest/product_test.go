package rest_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type fakeProductService struct {
	onAddProduct    func(ctx context.Context, request *menu.AddProductRequest) (*menu.Product, error)
	onUpdateProduct func(ctx context.Context, request *menu.UpdateProductRequest) error
	onUpdateImage   func(ctx context.Context, request *menu.UpdateImageRequest) (string, error)
	onDeleteProduct func(ctx context.Context, id uuid.UUID) error
	onGetProducts   func(ctx context.Context, filter *menu.ProductFilter) ([]menu.Product, error)
}

var _ menu.ProductService = (*fakeProductService)(nil)

func (s *fakeProductService) AddProduct(ctx context.Context, request *menu.AddProductRequest) (*menu.Product, error) {
	if s.onAddProduct == nil {
		panic("onAddProduct function is not implemented")
	}
	return s.onAddProduct(ctx, request)
}

func (s *fakeProductService) UpdateProduct(ctx context.Context, request *menu.UpdateProductRequest) error {
	if s.onUpdateProduct == nil {
		panic("onUpdateProduct function is not implemented")
	}
	return s.onUpdateProduct(ctx, request)
}

func (s *fakeProductService) UpdateImage(ctx context.Context, request *menu.UpdateImageRequest) (string, error) {
	if s.onUpdateImage == nil {
		panic("onUpdateImage function is not implemented")
	}
	return s.onUpdateImage(ctx, request)
}

func (s *fakeProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	if s.onDeleteProduct == nil {
		panic("onDeleteProduct function is not implemented")
	}
	return s.onDeleteProduct(ctx, id)
}

func (s *fakeProductService) GetProducts(ctx context.Context, filter *menu.ProductFilter) ([]menu.Product, error) {
	if s.onGetProducts == nil {
		panic("onGetProducts function is not implemented")
	}
	return s.onGetProducts(ctx, filter)
}

//go:embed testdata/product_image.jpg
var validProductImage []byte

func checkAddProductResponse(
	t *testing.T,
	body io.Reader,
	request *rest.AddProductRequest,
) {
	t.Helper()

	var res rest.ProductResponse
	if err := json.NewDecoder(body).Decode(&res); err != nil {
		t.Fatalf("failed to decode res: %v", err)
	}
	if request.Name != res.Name {
		t.Errorf("want name %s, got %s", request.Name, res.Name)
	}
	if request.Description != res.Description {
		t.Errorf("want description %s, got %s", request.Description, res.Description)
	}
	if !request.Price.Equal(res.Price) {
		t.Errorf("want price %s, got %s", request.Price, res.Price)
	}
	if request.CategoryId != res.CategoryId {
		t.Errorf("want category id %s, got %s", request.CategoryId, res.CategoryId)
	}
}

func creatAddProductRequest(t *testing.T, product *rest.AddProductRequest, image []byte) *http.Request {
	t.Helper()

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	productData, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("error encoding product: %v", err)
	}
	if err = writer.WriteField("product", string(productData)); err != nil {
		t.Fatalf("error writing product to multipart: %v", err)
	}

	imageWriter, err := writer.CreateFormFile("image", "image.png")
	if err != nil {
		t.Fatalf("error creating image writer: %v", err)
	}

	if _, err = imageWriter.Write(image); err != nil {
		t.Fatalf("error writing image: %v", err)
	}
	if err = writer.Close(); err != nil {
		t.Fatalf("error closing multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/products", &buffer)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}

func TestProductHandlerAddProduct(t *testing.T) {

	tests := []struct {
		name             string
		service          *fakeProductService
		productRequest   *rest.AddProductRequest
		image            []byte
		wantHttpStatus   int
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeProductService{
				onAddProduct: func(ctx context.Context, request *menu.AddProductRequest) (*menu.Product, error) {
					return &menu.Product{
						Id:          uuid.New(),
						Name:        request.Name,
						Description: request.Description,
						Price:       request.Price,
						CategoryId:  request.CategoryId,
						ImagePath:   "image/path",
					}, nil
				},
			},
			productRequest: &rest.AddProductRequest{
				Name:        "Valida product name",
				Description: "Valid product description",
				Price:       decimal.NewFromFloat(10.5),
				CategoryId:  uuid.New(),
			},
			image:          validProductImage,
			wantHttpStatus: http.StatusCreated,
		},
		{
			name:    "error invalid product",
			service: &fakeProductService{},
			productRequest: &rest.AddProductRequest{
				Name:        "",
				Description: "",
				Price:       decimal.NewFromFloat(-10.5),
				CategoryId:  uuid.New(),
			},
			image:          validProductImage,
			wantHttpStatus: http.StatusUnprocessableEntity,
			wantErrorCode:  domain.ErrorCodeInvalidProduct,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeProductNameTooShort,
				domain.ErrorCodeProductDescriptionTooShort,
				domain.ErrorCodeProductPriceLessThanZero,
			},
		},
		{
			name:    "error invalid product image",
			service: &fakeProductService{},
			productRequest: &rest.AddProductRequest{
				Name:        "Valida product name",
				Description: "Valid product description",
				Price:       decimal.NewFromFloat(10.5),
				CategoryId:  uuid.New(),
			},
			image:          []byte("invalid product image"),
			wantHttpStatus: http.StatusUnprocessableEntity,
			wantErrorCode:  domain.ErrorCodeInvalidProduct,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeInvalidImageType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := test.NewProductRouter(tt.service)
			request := creatAddProductRequest(t, tt.productRequest, tt.image)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantHttpStatus {
				t.Fatalf("want http status %d, got %d", tt.wantHttpStatus, recorder.Code)
			}

			if tt.wantHttpStatus == http.StatusCreated {
				checkAddProductResponse(t, recorder.Body, tt.productRequest)
			} else {
				test.CheckErrorResponse(t, recorder.Body, tt.wantErrorCode, tt.wantDetailsCodes...)
			}
		})
	}
}

func TestProductHandlerUpdateProduct(t *testing.T) {
	tests := []struct {
		name             string
		service          *fakeProductService
		id               string
		productRequest   *rest.UpdateProductRequest
		wantHttpStatus   int
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeProductService{
				onUpdateProduct: func(ctx context.Context, request *menu.UpdateProductRequest) error {
					return nil
				},
			},
			id: uuid.NewString(),
			productRequest: &rest.UpdateProductRequest{
				Name: new("new name"),
			},
			wantHttpStatus: http.StatusNoContent,
		},
		{
			name:           "no data",
			id:             uuid.NewString(),
			productRequest: &rest.UpdateProductRequest{},
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeNoData,
		},
		{
			name: "invalid update",
			id:   uuid.NewString(),
			productRequest: &rest.UpdateProductRequest{
				Name: new("na"),
			},
			wantHttpStatus: http.StatusUnprocessableEntity,
			wantErrorCode:  domain.ErrorCodeInvalidProductUpdate,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeProductNameTooShort,
			},
		},
		{
			name: "invalid id",
			id:   "invalid",
			productRequest: &rest.UpdateProductRequest{
				Name: new("New name"),
			},
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			body, err := json.Marshal(tt.productRequest)
			if err != nil {
				t.Fatalf("error encoding product: %v", err)
			}
			request := httptest.NewRequest(http.MethodPatch, "/products/"+tt.id, bytes.NewReader(body))
			router := test.NewProductRouter(tt.service)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantHttpStatus {
				t.Fatalf("want http status %d, got %d", tt.wantHttpStatus, recorder.Code)
			}
			if tt.wantHttpStatus == http.StatusNoContent {
				return
			}

			test.CheckErrorResponse(t, recorder.Body, tt.wantErrorCode, tt.wantDetailsCodes...)
		})
	}
}

func TestProductHandlerUpdateImage(t *testing.T) {
	tests := []struct {
		name             string
		service          *fakeProductService
		id               string
		image            []byte
		wantHttpStatus   int
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeProductService{
				onUpdateImage: func(ctx context.Context, request *menu.UpdateImageRequest) (string, error) {
					return "new/path", nil
				},
			},
			id:             uuid.NewString(),
			image:          validProductImage,
			wantHttpStatus: http.StatusCreated,
		},
		{
			name:           "invalid image",
			id:             uuid.NewString(),
			image:          []byte("invalid image"),
			wantHttpStatus: http.StatusUnprocessableEntity,
			wantErrorCode:  domain.ErrorCodeInvalidImageUpdate,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeInvalidImageType,
			},
		},
		{
			name:           "invalid id",
			id:             "invalid",
			image:          validProductImage,
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := test.NewProductRouter(tt.service)
			request := httptest.NewRequest(http.MethodPut, "/products/"+tt.id+"/image", bytes.NewReader(tt.image))
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantHttpStatus {
				t.Fatalf("want http status %d, got %d", tt.wantHttpStatus, recorder.Code)
			}

			if tt.wantHttpStatus == http.StatusCreated {
				var response rest.UpdateImageResponse
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
			} else {
				test.CheckErrorResponse(t, recorder.Body, tt.wantErrorCode, tt.wantDetailsCodes...)
			}
		})
	}
}

func TestProductHandlerDeleteProduct(t *testing.T) {
	tests := []struct {
		name           string
		service        *fakeProductService
		id             string
		wantHttpStatus int
		wantErrorCode  domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeProductService{
				onDeleteProduct: func(ctx context.Context, id uuid.UUID) error {
					return nil
				},
			},
			id:             uuid.NewString(),
			wantHttpStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid",
			id:             "invalid",
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := test.NewProductRouter(tt.service)
			request := httptest.NewRequest(http.MethodDelete, "/products/"+tt.id, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != tt.wantHttpStatus {
				t.Fatalf("want http status %d, got %d", tt.wantHttpStatus, recorder.Code)
			}

			if tt.wantHttpStatus == http.StatusOK {
				return
			}
			test.CheckErrorResponse(t, recorder.Body, tt.wantErrorCode)
		})
	}
}

func createGetProductsRequest(id *string, categoryId *string) *http.Request {
	var query = url.Values{}
	if id != nil {
		query.Add("id", *id)
	}
	if categoryId != nil {
		query.Add("category", *categoryId)
	}

	return httptest.NewRequest(http.MethodGet, "/products?"+query.Encode(), nil)
}

func TestProductHandlerGetProducts(t *testing.T) {
	tests := []struct {
		name           string
		service        *fakeProductService
		id             *string
		categoryId     *string
		wantHttpStatus int
		wantErrorCode  domain.ErrorCode
	}{
		{
			name: "success",
			service: &fakeProductService{
				onGetProducts: func(ctx context.Context, filter *menu.ProductFilter) ([]menu.Product, error) {
					return []menu.Product{}, nil
				},
			},
			id:             nil,
			categoryId:     nil,
			wantHttpStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             nil,
			categoryId:     new(""),
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := test.NewProductRouter(tt.service)
			request := createGetProductsRequest(tt.id, tt.categoryId)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)
			if recorder.Code != tt.wantHttpStatus {
				t.Fatalf("want http status %d, got %d", tt.wantHttpStatus, recorder.Code)
			}
			if tt.wantHttpStatus == http.StatusOK {
				var response []rest.ProductResponse
				if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
					t.Fatalf("error decoding response: %v", err)
				}
			} else {
				test.CheckErrorResponse(t, recorder.Body, tt.wantErrorCode)
			}
		})
	}
}
