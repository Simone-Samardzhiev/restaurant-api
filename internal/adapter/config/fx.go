package config

import "go.uber.org/fx"

var Module = fx.Module(
	"config",
	fx.Provide(New),
	fx.Provide(func(container *Container) *AppConfig {
		return &container.AppConfig
	}),
	fx.Provide(func(container *Container) *DBConfig {
		return &container.DbConfig
	}),
	fx.Provide(func(container *Container) *AuthConfig {
		return &container.AuthConfig
	}),
)
