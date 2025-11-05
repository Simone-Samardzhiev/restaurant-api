package validation

import (
	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"validator",
	fx.Provide(validator.New),
	fx.Invoke(func(v *validator.Validate) error {
		err := v.RegisterValidation("gtZero", func(fl validator.FieldLevel) bool {
			price, ok := fl.Field().Interface().(decimal.Decimal)
			if !ok {
				return false
			}

			return price.GreaterThan(decimal.Zero)
		},
		)

		return err
	}),
)
