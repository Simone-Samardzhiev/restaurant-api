package storage

import (
	"restaurant/internal/adapter/storage/http"
	"restaurant/internal/adapter/storage/postgres"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"storage",
	http.Module,
	postgres.Module,
)
