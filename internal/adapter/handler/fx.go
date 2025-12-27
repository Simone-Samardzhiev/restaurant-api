package handler

import (
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module(
	"handler",
	fx.Provide(NewRouter),
	fx.Invoke(func(lc fx.Lifecycle, router *Router) {
		lc.Append(fx.Hook{
			OnStart: func(context.Context) error {
				go func() {
					if err := router.Listen(); err != nil {
						zap.L().Error("http listen failed", zap.Error(err))
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return router.Shutdown(ctx)
			},
		})
	}),
)
