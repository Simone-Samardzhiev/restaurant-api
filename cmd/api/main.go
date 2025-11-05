package main

import (
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler/http"
	"restaurant/internal/adapter/handler/http/validation"
	"restaurant/internal/adapter/logger"
	"restaurant/internal/adapter/storage"
	"restaurant/internal/core/service"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		logger.Module,
		storage.Module,
		service.Module,
		validation.Module,
		http.Module,
	).Run()
}
