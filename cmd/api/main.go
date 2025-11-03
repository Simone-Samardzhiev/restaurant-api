package main

import (
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler/http"
	"restaurant/internal/adapter/logger"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		logger.Module,
		http.Module,
	).Run()
}
