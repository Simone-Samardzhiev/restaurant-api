package http

import (
	"restaurant/internal/adapter/storage/http/repository"
	"restaurant/internal/core/port"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"httpStorage",
	fx.Provide(
		fx.Annotate(
			repository.NewImageRepository,
			fx.As(new(port.ImageRepository)),
		),
	),
)
