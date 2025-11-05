package postgres

import (
	"restaurant/internal/adapter/storage/postgres/repository"
	"restaurant/internal/core/port"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"postgresStorage",
	fx.Provide(New),
	fx.Provide(
		fx.Annotate(
			repository.NewProductRepository,
			fx.As(new(port.ProductRepository)),
		),
	),
)
