package service

import (
	"restaurant/internal/core/port"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"service",
	fx.Provide(
		fx.Annotate(
			NewProductService,
			fx.As(new(port.ProductService)),
		),
	),
	fx.Provide(
		fx.Annotate(
			NewOrderService,
			fx.As(new(port.OrderService)),
		),
	),
)
