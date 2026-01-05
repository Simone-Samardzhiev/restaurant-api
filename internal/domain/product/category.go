package product

import (
	"fmt"
	"restaurant/internal/domain"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	MinCategoryNameLength = 4
	MaxCategoryNameLength = 100
)

// CategoryName represents a valid category name.
type CategoryName struct {
	raw string
}

// NewCategoryName parses a CategoryName from string.
func NewCategoryName(raw string) (CategoryName, error) {
	raw = strings.TrimSpace(raw)
	count := utf8.RuneCountInString(raw)

	if count < MinCategoryNameLength {
		return CategoryName{}, fmt.Errorf("category name must be at least %d characters", MinCategoryNameLength)
	}

	if count > MaxCategoryNameLength {
		return CategoryName{}, fmt.Errorf("category name must be at most %d characters", MaxCategoryNameLength)
	}

	return CategoryName{raw}, nil
}

func (c CategoryName) Raw() string {
	return c.raw
}

// AddCategoryRequest represents a request for adding a category.
type AddCategoryRequest struct {
	Name string
}

// NewAddCategoryRequest creates a new AddCategoryRequest.
func NewAddCategoryRequest(name string) *AddCategoryRequest {
	return &AddCategoryRequest{Name: name}
}

// Category represents a category entity.
type Category struct {
	Id   uuid.UUID
	Name CategoryName
}

// NewCategory creates a Category by parsing the name.
func NewCategory(id uuid.UUID, name string) (*Category, error) {
	validationErrors := domain.NewValidationErrors("invalid category")

	parsedName, err := NewCategoryName(name)
	if err != nil {
		validationErrors.Add("name", err)
	}

	if validationErrors.HasErrors() {
		return nil, validationErrors
	}

	return &Category{id, parsedName}, nil
}

func (c *Category) RawName() string {
	return c.Name.raw
}
