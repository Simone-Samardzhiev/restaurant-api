package validator

import (
	"restaurant/internal/core/domain"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

func validatePrice(fl validator.FieldLevel) bool {
	price, ok := fl.Field().Interface().(decimal.Decimal)
	if !ok {
		return false
	}

	return price.GreaterThan(decimal.Zero)
}

var orderStatuses = map[domain.OrderSessionStatus]struct{}{
	domain.Closed: {},
	domain.Open:   {},
	domain.Paid:   {},
}

func validateOrderSessionStatus(fl validator.FieldLevel) bool {
	status, ok := fl.Field().Interface().(domain.OrderSessionStatus)
	if !ok {
		return false
	}
	_, exists := orderStatuses[status]
	return exists
}

var orderedProductStatuses = map[domain.OrderedProductStatus]struct{}{
	domain.Pending:   {},
	domain.Preparing: {},
	domain.Done:      {},
}

func validateOrderedProductStatus(fl validator.FieldLevel) bool {
	status, ok := fl.Field().Interface().(domain.OrderedProductStatus)
	if !ok {
		return false
	}
	_, exists := orderedProductStatuses[status]
	return exists
}
