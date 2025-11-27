package websocket

import "go.uber.org/fx"

var Module = fx.Module(
	"websocket",
	fx.Provide(NewHub),
	fx.Invoke(func(hub *Hub) {
		go hub.Run()
	}),
	fx.Provide(NewHandler),
)
