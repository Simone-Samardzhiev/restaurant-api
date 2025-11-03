package http

import (
	"context"

	"go.uber.org/fx"
)

var Module = fx.Module(
	"http",
	fx.Provide(NewRouter),
	fx.Invoke(func(lc fx.Lifecycle, router *Router) {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				return router.Listen()
			},
			OnStop: func(context.Context) error {
				return router.Shutdown()
			},
		})
	}),
)
