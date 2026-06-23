package domain

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type fakeCategoryRepository struct {
	onSave   func(ctx context.Context, category *Category) error
	onGetAll func(ctx context.Context) ([]Category, error)
	onUpdate func(ctx context.Context, id uuid.UUID, name string) error
	onDelete func(ctx context.Context, id uuid.UUID) error
}

var _ CategoryRepository = (*fakeCategoryRepository)(nil)

func (f fakeCategoryRepository) Save(ctx context.Context, category *Category) error {
	if f.onSave == nil {
		panic("onSave not implemented")
	}
	return f.onSave(ctx, category)
}

func (f fakeCategoryRepository) GetAll(ctx context.Context) ([]Category, error) {
	if f.onGetAll == nil {
		panic("onGetAll not implemented")
	}
	return f.onGetAll(ctx)
}

func (f fakeCategoryRepository) Update(ctx context.Context, id uuid.UUID, name string) error {
	if f.onUpdate == nil {
		panic("onUpdate not implemented")
	}
	return f.onUpdate(ctx, id, name)
}

func (f fakeCategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if f.onDelete == nil {
		panic("onDelete not implemented")
	}
	return f.onDelete(ctx, id)
}

func TestDefaultMenuServiceGet(t *testing.T) {
	tests := []struct {
		categoryRepository *fakeCategoryRepository
		productRepository  *fakeProductRepository
		want               Menu
	}{
		{
			categoryRepository: &fakeCategoryRepository{
				onGetAll: func(ctx context.Context) ([]Category, error) {
					return []Category{
						{
							Id:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
							Name:      "Category 1",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
						{
							Id:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
							Name:      "Category 2",
							CreatedAt: time.Now(),
							UpdatedAt: time.Now(),
						},
					}, nil
				},
			},
			productRepository: &fakeProductRepository{
				onGetAllWithImage: func(ctx context.Context) ([]Product, error) {
					return []Product{
						{
							Id:               uuid.New(),
							Name:             "Product 1",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
						{
							Id:               uuid.New(),
							Name:             "Product 2",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
						{
							Id:               uuid.New(),
							Name:             "Product 3",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
						{
							Id:               uuid.New(),
							Name:             "Product 4",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
						{
							Id:               uuid.New(),
							Name:             "Product 5",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
						{
							Id:               uuid.New(),
							Name:             "Product 6",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
					}, nil
				},
			},
			want: Menu{
				MenuSection{
					Category: Category{
						Id:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
						Name:      "Category 1",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Products: []Product{
						{
							Id:               uuid.New(),
							Name:             "Product 1",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
						{
							Id:               uuid.New(),
							Name:             "Product 2",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
						{
							Id:               uuid.New(),
							Name:             "Product 4",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("11111111-1111-1111-1111-111111111111"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
					},
				},
				MenuSection{
					Category: Category{
						Id:        uuid.MustParse("22222222-2222-2222-2222-222222222222"),
						Name:      "Category 2",
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
					Products: []Product{
						{
							Id:               uuid.New(),
							Name:             "Product 3",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
						{
							Id:               uuid.New(),
							Name:             "Product 5",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
						{
							Id:               uuid.New(),
							Name:             "Product 6",
							Description:      "Some test description",
							Price:            decimal.NewFromInt(100),
							CategoryId:       uuid.MustParse("22222222-2222-2222-2222-222222222222"),
							ImageKey:         "imageKey",
							ImageContentType: ImageContentTypePNG,
							Status:           ProductStatusReady,
							CreatedAt:        time.Now(),
							UpdatedAt:        time.Now(),
						},
					},
				},
			},
		},
	}

	for i, tt := range tests {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			t.Parallel()

			service := NewDefaultMenuService(tt.categoryRepository, tt.productRepository)
			result, err := service.Get(context.Background())
			if err != nil {
				t.Fatalf("error getting menu: %v", err)
			}

			slices.SortFunc(result, func(a, b MenuSection) int {
				return strings.Compare(a.Category.Name, b.Category.Name)
			})
			for _, section := range result {
				slices.SortFunc(section.Products, func(a, b Product) int {
					return strings.Compare(a.Name, b.Name)
				})
			}

			if len(result) != len(tt.want) {
				t.Fatalf("got %v sections, want: %v", len(result), len(tt.want))
			}
			for j := 0; j < len(result); j++ {
				if result[j].Category.Name != tt.want[j].Category.Name {
					t.Errorf("got category name: %v, want: %v", result[j].Category.Name, tt.want[j].Category.Name)
				}
			}

			for j := 0; j < len(result); j++ {
				if len(result[j].Products) != len(tt.want[j].Products) {
					t.Errorf("got %v products, want: %v", result[j].Products, tt.want[j].Products)
					continue
				}

				for k := 0; k < len(result[j].Products); k++ {
					if result[j].Products[k].Name != tt.want[j].Products[k].Name {
						t.Errorf("got product name: %v, want: %v", result[j].Products[k].Name, tt.want[j].Products[k].Name)
					}
				}
			}
		})
	}
}
