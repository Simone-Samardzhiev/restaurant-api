package rest

import (
	"context"
	"encoding/json"
	"menu/internal/database"
	"menu/internal/domain"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"
)

func TestGetMenu(t *testing.T) {
	categoryRepository := database.NewPostgresCategoryRepository(testDb)
	productRepository := database.NewPostgresProductRepository(testDb)
	menuService := domain.NewDefaultMenuService(categoryRepository, productRepository)
	menuHandler := NewMenuHandler("http://images", menuService)
	e := echo.NewWithConfig(echo.Config{
		HTTPErrorHandler: ErrorHandler,
	})
	e.GET("/menu", menuHandler.GetMenu)

	if _, err := testDb.Exec(`TRUNCATE TABLE categories, products CASCADE`); err != nil {
		t.Fatalf("error truncating table: %v", err)
	}

	menu := domain.Menu{
		domain.MenuSection{
			Category: domain.Category{
				Id:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				Name:      "Category 1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Products: []domain.Product{
				{
					Id:               uuid.New(),
					Name:             "Product 1",
					Description:      "Some test description",
					Price:            decimal.NewFromInt(100),
					CategoryId:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
					ImageKey:         "imageKey1",
					ImageContentType: domain.ImageContentTypePNG,
					Status:           domain.ProductStatusReady,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
			},
		},
		domain.MenuSection{
			Category: domain.Category{
				Id:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				Name:      "Category 2",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			Products: []domain.Product{
				{
					Id:               uuid.UUID{},
					Name:             "Product 2",
					Description:      "Some test description",
					Price:            decimal.NewFromInt(100),
					CategoryId:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
					ImageKey:         "imageKey2",
					ImageContentType: domain.ImageContentTypePNG,
					Status:           domain.ProductStatusReady,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
			},
		},
	}

	for _, section := range menu {
		if err := categoryRepository.Save(context.Background(), &section.Category); err != nil {
			t.Fatalf("error saving category: %v", err)
		}
		for _, product := range section.Products {
			if err := productRepository.Save(context.Background(), &product); err != nil {
				t.Fatalf("error saving product: %v", err)
			}
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/menu", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want http status: %d, got: %d", http.StatusOK, rec.Code)
	}

	var res []MenuSectionResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("error decoding response: %v", err)
	}

	slices.SortFunc(res, func(a, b MenuSectionResponse) int {
		return strings.Compare(a.Name, b.Name)
	})
	for _, section := range res {
		slices.SortFunc(section.Products, func(a, b ProductResponse) int {
			return strings.Compare(a.Name, b.Name)
		})
	}

	if len(res) != len(menu) {
		t.Fatalf("got %v sections, want: %v", len(res), len(menu))
	}
	for j := 0; j < len(res); j++ {
		if res[j].CategoryResponse.Name != menu[j].Category.Name {
			t.Errorf("got category name: %v, want: %v", res[j].CategoryResponse.Name, menu[j].Category.Name)
		}
	}

	for j := 0; j < len(res); j++ {
		if len(res[j].Products) != len(menu[j].Products) {
			t.Errorf("got %v products, want: %v", res[j].Products, menu[j].Products)
			continue
		}

		for k := 0; k < len(res[j].Products); k++ {
			if res[j].Products[k].Name != menu[j].Products[k].Name {
				t.Errorf("got product name: %v, want: %v", res[j].Products[k].Name, menu[j].Products[k].Name)
			}
		}
	}

}
