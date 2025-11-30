package validation

import (
	"restaurant/internal/adapter/handler/websocket"
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

var orderedProductStatuses = map[domain.OrderedProductStatus]bool{
	domain.Pending:   true,
	domain.Preparing: true,
	domain.Done:      true,
}

func validateOrderedProductStatus(fl validator.FieldLevel) bool {
	status, ok := fl.Field().Interface().(domain.OrderedProductStatus)
	if !ok {
		return false
	}
	return orderedProductStatuses[status]
}

var messageTypes = map[websocket.MessageType]bool{
	websocket.Order: true,
}

func validateMessageType(fl validator.FieldLevel) bool {
	messageType, ok := fl.Field().Interface().(websocket.MessageType)
	if !ok {
		return false
	}
	return messageTypes[messageType]
}

var Module = fx.Module(
	"validator",
	fx.Provide(validator.New),
	fx.Invoke(func(v *validator.Validate) error {
		if err := v.RegisterValidation("gtZero", validatePrice); err != nil {
			return err
		}
		if err := v.RegisterValidation("orderStatus", validateOrderSessionStatus); err != nil {
			return err
		}
		if err := v.RegisterValidation("orderedProductStatus", validateOrderedProductStatus); err != nil {
			return err
		}
		if err := v.RegisterValidation("messageType", validateMessageType); err != nil {
			return err
		}

		return nil
	}),
)
