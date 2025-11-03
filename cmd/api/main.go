package main

import (
	"restaurant/internal/adapter/config"
	"restaurant/internal/adapter/handler/http"

	"go.uber.org/fx"
)

func main() {
	fx.New(
		config.Module,
		http.Module,
	).Run()
}
