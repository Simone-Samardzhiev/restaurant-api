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
	"restaurant/internal/adapter/handler/rest/middleware"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
	"testing"

	"github.com/gin-gonic/gin"
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

// createAddCategoryRouter creates a gin router with /product path for [ProductHandler.AddProduct].
func createAddProductRouter(service menu.ProductService) *gin.Engine {
	handler := rest.NewProductHandler(service, "")
	router := gin.New()
	router.Use(middleware.Error())
	router.POST("/product", handler.AddProduct)
	return router
}

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

	request := httptest.NewRequest(http.MethodPost, "/product", &buffer)
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

			router := createAddProductRouter(tt.service)
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

// createUpdateProductRouter creates a gin router with /product/:id path for
// [ProductHandler.UpdateProduct].
func createUpdateProductRouter(service menu.ProductService) *gin.Engine {
	handler := rest.NewProductHandler(service, "")
	router := gin.New()
	router.Use(middleware.Error())
	router.PATCH("/product/:id", handler.UpdateProduct)
	return router
}

func createUpdateProductRequest(t *testing.T, id uuid.UUID, productRequest *rest.UpdateProductRequest) *http.Request {
	t.Helper()

	body, err := json.Marshal(productRequest)
	if err != nil {
		t.Fatalf("error encoding body: %v", err)
	}

	return httptest.NewRequest(http.MethodPatch, "/product/"+id.String(), bytes.NewBuffer(body))
}

func TestProductHandlerUpdateProduct(t *testing.T) {
	tests := []struct {
		name             string
		service          *fakeProductService
		id               uuid.UUID
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
			id: uuid.New(),
			productRequest: &rest.UpdateProductRequest{
				Name: new("new name"),
			},
			wantHttpStatus: http.StatusNoContent,
		},
		{
			name:           "no data",
			id:             uuid.New(),
			productRequest: &rest.UpdateProductRequest{},
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeNoData,
		},
		{
			name: "invalid update",
			id:   uuid.New(),
			productRequest: &rest.UpdateProductRequest{
				Name: new("na"),
			},
			wantHttpStatus: http.StatusUnprocessableEntity,
			wantErrorCode:  domain.ErrorCodeInvalidProductUpdate,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeProductNameTooShort,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := createUpdateProductRouter(tt.service)
			request := createUpdateProductRequest(t, tt.id, tt.productRequest)
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

// creteUpdateImageRouter creates a gin router with PUT /:id/image
// for [ProductHandler.UpdateImage].
func creteUpdateImageRouter(service menu.ProductService) *gin.Engine {
	handler := rest.NewProductHandler(service, "")
	router := gin.New()
	router.Use(middleware.Error())
	router.PUT("/:id/image", handler.UpdateImage)
	return router
}

func TestProductHandlerUpdateImage(t *testing.T) {
	tests := []struct {
		name             string
		service          *fakeProductService
		id               uuid.UUID
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
			id:             uuid.New(),
			image:          validProductImage,
			wantHttpStatus: http.StatusCreated,
		},
		{
			name:           "invalid request",
			id:             uuid.New(),
			image:          []byte("invalid image"),
			wantHttpStatus: http.StatusUnprocessableEntity,
			wantErrorCode:  domain.ErrorCodeInvalidImageUpdate,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeInvalidImageType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := creteUpdateImageRouter(tt.service)
			request := httptest.NewRequest(http.MethodPut, "/"+tt.id.String()+"/image", bytes.NewReader(tt.image))
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

// createDeleteProductRouter creates a new gin router with DELETE product/:id
// for [ProductHandler.DeleteProduct].
func createDeleteProductRouter(service menu.ProductService) *gin.Engine {
	handler := rest.NewProductHandler(service, "")
	router := gin.New()
	router.Use(middleware.Error())
	router.DELETE("/product/:id", handler.DeleteProduct)
	return router
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

			router := createDeleteProductRouter(tt.service)
			request := httptest.NewRequest(http.MethodDelete, "/product/"+tt.id, nil)
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

// createGetProductsRouter creates a gin router with GET /product for
// [ProductHandler.GetProducts].
func createGetProductsRouter(service menu.ProductService) *gin.Engine {
	handler := rest.NewProductHandler(service, "")
	router := gin.New()
	router.Use(middleware.Error())
	router.GET("/product", handler.GetProducts)
	return router
}

func createGetProductsRequest(id *string, categoryId *string) *http.Request {
	var query = url.Values{}
	if id != nil {
		query.Add("id", *id)
	}
	if categoryId != nil {
		query.Add("category", *categoryId)
	}

	return httptest.NewRequest(http.MethodGet, "/product?"+query.Encode(), nil)
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
			name:           "invalid request",
			id:             nil,
			categoryId:     new(""),
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := createGetProductsRouter(tt.service)
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
