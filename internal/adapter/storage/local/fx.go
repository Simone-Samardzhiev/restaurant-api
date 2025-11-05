package local

import (
	"restaurant/internal/adapter/storage/local/repository"
	"restaurant/internal/core/port"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"localStorage",
	fx.Provide(
		fx.Annotate(
			repository.NewImageRepository,
			fx.As(new(port.ImageRepository)),
		),
	),
)
