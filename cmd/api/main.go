package main

import (
	"fmt"
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

	for i := 10; i < 20; i++ {
		fmt.Print(i)
	}
}
