package main

import (
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler"
	"restaurant/internal/adapter/handler/http"
	"restaurant/internal/adapter/handler/http/validator"
	"restaurant/internal/adapter/handler/websocket"
	"restaurant/internal/adapter/logger"
	"restaurant/internal/adapter/storage"
	"restaurant/internal/core/service"
	"time"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.StopTimeout(30*time.Second),
		config.Module,
		logger.Module,
		storage.Module,
		service.Module,
		validator.Module,
		http.Module,
		websocket.Module,
		handler.Module,
	).Run()
}
