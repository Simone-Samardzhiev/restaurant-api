package logger

import (
	"restaurant/internal/adapter/config"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"logger",
	fx.Invoke(func(appConfig *config.AppConfig) error {
		return setZapLogger(appConfig)
	}),
)
