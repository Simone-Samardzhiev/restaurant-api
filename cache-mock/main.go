package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
)

func main() {
	http.HandleFunc("/zones/{id}/purge_cache", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
		log.Println("Cache purge request received")
	})

	go func() {
		_ = http.ListenAndServe(":8080", nil)
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	<-ctx.Done()

	log.Println("Shutting down")
}
