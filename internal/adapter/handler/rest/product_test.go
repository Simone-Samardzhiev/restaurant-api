package rest_test

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/handler/rest/middleware"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type fakeProductService struct {
	onAddProduct    func(ctx context.Context, request *menu.AddProductRequest) (*menu.Product, error)
	onUpdateProduct func(ctx context.Context, request *menu.UpdateProductRequest) error
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
	body []byte,
	wantName,
	wantDescription string,
	wantPrice decimal.Decimal,
	wantCategoryId uuid.UUID,
) {
	t.Helper()

	var response rest.ProductResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if wantName != response.Name {
		t.Errorf("want name %s, got %s", wantName, response.Name)
	}
	if wantDescription != response.Description {
		t.Errorf("want description %s, got %s", wantDescription, response.Description)
	}
	if !wantPrice.Equal(response.Price) {
		t.Errorf("want price %s, got %s", wantPrice, response.Price)
	}
	if wantCategoryId != response.CategoryId {
		t.Errorf("want categoryId %s, got %s", wantCategoryId, response.CategoryId)
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
	gin.SetMode(gin.TestMode)

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
				CategoryID:  uuid.New(),
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
				CategoryID:  uuid.New(),
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
				CategoryID:  uuid.New(),
			},
			image:          []byte("invalid product image"),
			wantHttpStatus: http.StatusUnprocessableEntity,
			wantErrorCode:  domain.ErrorCodeInvalidProduct,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeInvalidImageType,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			router := createAddProductRouter(test.service)
			request := creatAddProductRequest(t, test.productRequest, test.image)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != test.wantHttpStatus {
				t.Fatalf("want http status %d, got %d", test.wantHttpStatus, recorder.Code)
			}

			if test.wantHttpStatus == http.StatusCreated {
				checkAddProductResponse(
					t,
					recorder.Body.Bytes(),
					test.productRequest.Name,
					test.productRequest.Description,
					test.productRequest.Price,
					test.productRequest.CategoryID,
				)
			} else {
				checkErrorResponse(t, recorder.Body.Bytes(), test.wantErrorCode, test.wantDetailsCodes...)
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
	gin.SetMode(gin.TestMode)

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

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			router := createUpdateProductRouter(test.service)
			request := createUpdateProductRequest(t, test.id, test.productRequest)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != test.wantHttpStatus {
				t.Fatalf("want http status %d, got %d", test.wantHttpStatus, recorder.Code)
			}
			if test.wantHttpStatus == http.StatusNoContent {
				return
			}

			checkErrorResponse(t, recorder.Body.Bytes(), test.wantErrorCode, test.wantDetailsCodes...)
		})
	}
}
