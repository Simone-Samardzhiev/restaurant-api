package test

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/handler/rest/middleware"
	"restaurant/internal/adapter/handler/translator"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/domain/order"
	"testing"

	"github.com/gin-gonic/gin"
)

// CheckErrorResponse checks if the response is [middleware.ErrorResponse] and
// the error code and details code match the arguments
func CheckErrorResponse(
	t testing.TB,
	body io.Reader,
	expectedCode domain.ErrorCode,
	expectedCodes ...domain.ErrorCode,
) {
	t.Helper()

	var resp middleware.ErrorResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response body: %s", err)
	}

	if expectedCode.String() != resp.Code {
		t.Errorf("want error code %s, got %s", expectedCode.String(), resp.Code)
	}

	if len(expectedCodes) > 0 {
		matchErrorCodes(t, expectedCodes, resp.Details)
	}
}

// matchErrorCodes checks if error codes in match the error codes in the details, in any order.
func matchErrorCodes(t testing.TB, wantCodes []domain.ErrorCode, details []translator.ErrorResponseDetail) {
	t.Helper()

	if len(details) != len(details) {
		t.Fatalf("want %d codes, got %d ", len(wantCodes), len(details))
	}

	var counter = map[string]int{}
	for _, code := range wantCodes {
		counter[code.String()]++
	}

	for _, detail := range details {
		if counter[detail.Code] == 0 {
			t.Errorf("unexpected error code: %s", detail.Code)
			continue
		}
		counter[detail.Code]--
	}

	for code, count := range counter {
		if count != 0 {
			t.Errorf("missing error code: %s", code)
		}
	}
}

// NewCategoryRouter creates a new [gin.Engine] with routes for [rest.CategoryHandler].
//
// Path to each method:
//   - POST /categories - [rest.CategoryHandler.AddCategory].
//   - PATCH /categories/:id - [rest.CategoryHandler.UpdateCategory].
//   - DELETE /categories/:id - [rest.CategoryHandler.DeleteCategory].
//   - GET /categories - [rest.CategoryHandler.GetCategories].
func NewCategoryRouter(service menu.CategoryService) *gin.Engine {
	handler := rest.NewCategoryHandler(service)
	router := gin.New()
	router.Use(middleware.Error())
	router.POST("/categories", handler.AddCategory)
	router.PATCH("/categories/:id", handler.UpdateCategory)
	router.DELETE("/categories/:id", handler.DeleteCategory)
	router.GET("/categories", handler.GetCategories)
	return router
}

// NewCategoryRouterFromHandler creates a new [gin.Engine] with routes for the provided [rest.CategoryHandler].
//
// Path to each method:
//   - POST /categories - [rest.CategoryHandler.AddCategory].
//   - PATCH /categories/:id - [rest.CategoryHandler.UpdateCategory].
//   - DELETE /categories/:id - [rest.CategoryHandler.DeleteCategory].
//   - GET /categories - [rest.CategoryHandler.GetCategories].
func NewCategoryRouterFromHandler(handler *rest.CategoryHandler) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Error())
	router.POST("/categories", handler.AddCategory)
	router.PATCH("/categories/:id", handler.UpdateCategory)
	router.DELETE("/categories/:id", handler.DeleteCategory)
	router.GET("/categories", handler.GetCategories)
	return router
}

// NewProductRouter creates a new [gin.Engine] with routes for [rest.ProductHandler].
//
// Path to each method:
//   - POST /products - [rest.ProductHandler.AddProduct].
//   - PATCH /products/:id - [rest.ProductHandler.UpdateProduct]
//   - PUT /products/:id/image - [rest.ProductHandler.UpdateImage].
//   - DELETE /products/:id - [rest.ProductHandler.DeleteProduct]
//   - GET /products - [rest.ProductHandler.GetProducts].
func NewProductRouter(service menu.ProductService) *gin.Engine {
	handler := rest.NewProductHandler(service, "servingPath")
	router := gin.New()
	router.Use(middleware.Error())
	router.POST("/products", handler.AddProduct)
	router.PATCH("/products/:id", handler.UpdateProduct)
	router.PUT("/products/:id/image", handler.UpdateImage)
	router.DELETE("/products/:id", handler.DeleteProduct)
	router.GET("/products", handler.GetProducts)
	return router
}

// NewProductRouterFromHandler creates a new [gin.Engine] with routes for the provided [rest.ProductHandler].
//
// Path to each method:
//   - POST /products - [rest.ProductHandler.AddProduct].
//   - PATCH /products/:id - [rest.ProductHandler.UpdateProduct]
//   - PUT /products/:id/image - [rest.ProductHandler.UpdateImage].
//   - DELETE /products/:id - [rest.ProductHandler.DeleteProduct]
//   - GET /products - [rest.ProductHandler.GetProducts].
func NewProductRouterFromHandler(handler *rest.ProductHandler) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Error())
	router.POST("/products", handler.AddProduct)
	router.PATCH("/products/:id", handler.UpdateProduct)
	router.PUT("/products/:id/image", handler.UpdateImage)
	router.DELETE("/products/:id", handler.DeleteProduct)
	router.GET("/products", handler.GetProducts)
	return router
}

// NewAddProductRequest creates an http.Request for adding a product.
func NewAddProductRequest(t testing.TB, product *rest.AddProductRequest, image []byte) *http.Request {
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

// NewGetProductsRequest creates an http.Request for fetching the products.
func NewGetProductsRequest(id *string, categoryId *string) *http.Request {
	var query = url.Values{}
	if id != nil {
		query.Add("id", *id)
	}
	if categoryId != nil {
		query.Add("category", *categoryId)
	}

	return httptest.NewRequest(http.MethodGet, "/products?"+query.Encode(), nil)
}

// NewSessionRouter creates a new [gin.Engine] with router for [rest.SessionHandler].
//
// Path to each method:
//   - GET /sessions/:id - [rest.SessionHandler.GetSessionDetails]
func NewSessionRouter(service order.SessionService) *gin.Engine {
	handler := rest.NewSessionHandler(service)
	router := gin.New()
	router.Use(middleware.Error())
	router.GET("/sessions/:id", handler.GetSessionDetails)
	return router
}

// NewSessionRouterFromHandler creates a new [gin.Engine] with routes for the provided [rest.SessionHandler]
//
// Path to each method:
//   - GET /sessions/:id - [rest.SessionHandler.GetSessionDetails]
func NewSessionRouterFromHandler(handler *rest.SessionHandler) *gin.Engine {
	router := gin.New()
	router.Use(middleware.Error())
	router.GET("/sessions/:id", handler.GetSessionDetails)
	return router
}
