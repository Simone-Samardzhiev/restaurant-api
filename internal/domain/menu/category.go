package menu

import (
	"errors"
	"restaurant/internal/domain"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// Constants for [CategoryName] length requirements.
const (
	minCategoryNameLength = 4
	maxCategoryNameLength = 100
)

// CategoryName represents a valid category name.
type CategoryName struct {
	raw string
}

func (c *CategoryName) String() string {
	return c.raw
}

// ParseCategoryName parses a [CategoryName] from string.
//
// If the name is invalid the returned error will be of type [domain.ErrorDetail].
func ParseCategoryName(name string) (CategoryName, error) {
	name = strings.TrimSpace(name)
	length := utf8.RuneCountInString(name)
	if length < minCategoryNameLength {
		return CategoryName{}, &domain.ErrorDetail{
			Code:    domain.ErrorCodeCategoryNameTooShort,
			Message: "category name too short",
			Metadata: map[string]any{
				"actual": length,
				"min":    minCategoryNameLength,
				"max":    maxCategoryNameLength,
			},
		}
	}

	if length > maxCategoryNameLength {
		return CategoryName{}, &domain.ErrorDetail{
			Code:    domain.ErrorCodeCategoryNameTooLong,
			Message: "category name too long",
			Metadata: map[string]any{
				"actual": length,
				"max":    maxCategoryNameLength,
				"min":    minCategoryNameLength,
			},
		}
	}

	return CategoryName{raw: name}, nil
}

// Category represents a valid category of the menu.
type Category struct {
	Id   uuid.UUID
	Name CategoryName
}

// ParseCategory parses a [Category] from id and name.
//
// If the name is invalid the returned error will be of type [domain.Error].
func ParseCategory(id uuid.UUID, name string) (*Category, error) {
	errs := make([]domain.ErrorDetail, 0)

	parsedName, err := ParseCategoryName(name)
	if err != nil {
		if errorDetail, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *errorDetail)
		} else {
			return nil, err
		}
	}

	if len(errs) > 0 {
		return nil, domain.NewValidationError(
			"invalid category",
			domain.ErrorCodeInvalidCategory, errs...,
		)
	}

	return &Category{
		Id:   id,
		Name: parsedName,
	}, nil
}

// AddCategoryRequest represents a request for adding a new category.
type AddCategoryRequest struct {
	Name CategoryName
}

// ParseAddCategoryRequest parses an [AddCategoryRequest] from name.
//
// If the name is invalid the returned error will be of type [domain.Error].
func ParseAddCategoryRequest(name string) (*AddCategoryRequest, error) {
	errs := make([]domain.ErrorDetail, 0)

	parsedName, err := ParseCategoryName(name)
	if err != nil {
		if errorDetail, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *errorDetail)
		} else {
			return nil, err
		}
	}

	if len(errs) > 0 {
		return nil, domain.NewValidationError(
			"invalid add request for category",
			domain.ErrorCodeInvalidCategory,
			errs...,
		)
	}

	return &AddCategoryRequest{
		Name: parsedName,
	}, nil
}

// UpdateCategoryRequest represents a request for updating a category.
type UpdateCategoryRequest struct {
	Id   uuid.UUID
	Name *CategoryName
}

// ParseUpdateCategoryRequest parses [UpdateCategoryRequest] from id and name.
//
// If the name is invalid or the request update data is empty the error will be of type [domain.Error].
func ParseUpdateCategoryRequest(id uuid.UUID, name *string) (*UpdateCategoryRequest, error) {
	if name == nil {
		return nil, domain.NewBadRequestError("update does not have data", domain.ErrorCodeNoData, nil)
	}

	errs := make([]domain.ErrorDetail, 0)
	parsedName, err := ParseCategoryName(*name)
	if err != nil {
		if errorDetail, ok := errors.AsType[*domain.ErrorDetail](err); ok {
			errs = append(errs, *errorDetail)
		} else {
			return nil, err
		}
	}

	if len(errs) > 0 {
		return nil, domain.NewValidationError("invalid update request for category", domain.ErrorCodeInvalidCategoryUpdate, errs...)
	}

	return &UpdateCategoryRequest{
		Id:   id,
		Name: &parsedName,
	}, nil
}

// CategoryFilter represents a filter used for fetching categories.
type CategoryFilter struct {
	Id *uuid.UUID
}
