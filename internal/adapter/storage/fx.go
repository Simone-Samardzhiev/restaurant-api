package storage

import (
	"restaurant/internal/adapter/storage/local"
	"restaurant/internal/adapter/storage/postgres"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"storage",
	local.Module,
	postgres.Module,
)
