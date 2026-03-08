package test

import (
	"encoding/json"
	"io"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/adapter/handler/rest/middleware"
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"testing"

	"github.com/gin-gonic/gin"
)

// CheckErrorResponse checks if the response is [middleware.ErrorResponse] and
// the error code and details code match the arguments
func CheckErrorResponse(
	t *testing.T,
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
		MatchErrorCodes(t, expectedCodes, resp.Details)
	}
}

// MatchErrorCodes checks if error codes in match the error codes in the details, in any order.
func MatchErrorCodes(t *testing.T, wantCodes []domain.ErrorCode, details []middleware.ErrorDetailsResponse) {
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

// CreateCategoryRouter creates a new [gin.Engine] with routes for [rest.CategoryHandler].
//
// Path to each method:
//   - POST /categories - [rest.CategoryHandler.AddCategory].
//   - PATCH /categories/:id - [rest.CategoryHandler.UpdateCategory].
//   - DELETE /categories/:id - [rest.CategoryHandler.DeleteCategory].
//   - GET /categories - [rest.CategoryHandler.GetCategories].
func CreateCategoryRouter(service menu.CategoryService) *gin.Engine {
	handler := rest.NewCategoryHandler(service)
	router := gin.New()
	router.Use(middleware.Error())
	router.POST("/categories", handler.AddCategory)
	router.PATCH("/categories/:id", handler.UpdateCategory)
	router.DELETE("/categories/:id", handler.DeleteCategory)
	router.GET("/categories", handler.GetCategories)
	return router
}

// CreateProductRouter creates a new [gin.Engine] with routes for [rest.ProductHandler].
//
// Path to each method:
//   - POST /products - [rest.ProductHandler.AddProduct].
//   - PATCH /products/:id - [rest.ProductHandler.UpdateProduct]
//   - PUT /products/:id/image - [rest.ProductHandler.UpdateImage].
//   - DELETE /products/:id - [rest.ProductHandler.DeleteProduct]
//   - GET /products - [rest.ProductHandler.GetProducts].
func CreateProductRouter(service menu.ProductService) *gin.Engine {
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
