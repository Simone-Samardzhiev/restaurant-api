package websocket

import "go.uber.org/fx"

var Module = fx.Module(
	"websocket",
	fx.Provide(newHub),
	fx.Invoke(func(hub *hub) {
		go hub.Listen()
	}),
	fx.Provide(NewHandler),
)
