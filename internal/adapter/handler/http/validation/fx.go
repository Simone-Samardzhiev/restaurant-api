package validation

import (
	"restaurant/internal/core/domain"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"go.uber.org/fx"
)

func validatePrice(fl validator.FieldLevel) bool {
	price, ok := fl.Field().Interface().(decimal.Decimal)
	if !ok {
		return false
	}

	return price.GreaterThan(decimal.Zero)
}

var orderStatuses = map[domain.OrderSessionStatus]bool{
	domain.Closed: true,
	domain.Open:   true,
	domain.Paid:   true,
}

func validateOrderSessionStatus(fl validator.FieldLevel) bool {
	status, ok := fl.Field().Interface().(domain.OrderSessionStatus)
	if !ok {
		return false
	}
	return orderStatuses[status]
}

var Module = fx.Module(
	"validator",
	fx.Provide(validator.New),
	fx.Invoke(func(v *validator.Validate) error {
		err := v.RegisterValidation("gtZero", validatePrice)
		return err
	}),

	fx.Invoke(func(v *validator.Validate) error {
		err := v.RegisterValidation("orderStatus", validateOrderSessionStatus)
		return err
	}),
)
