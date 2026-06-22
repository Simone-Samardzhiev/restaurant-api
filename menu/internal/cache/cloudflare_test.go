package cache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/option"
)

func newTestServer() *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/zones/{id}/purge_cache", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	})

	return httptest.NewServer(mux)
}

func TestCloudflareCachePurgerCategories(t *testing.T) {
	server := newTestServer()
	t.Cleanup(server.Close)

	client := cloudflare.NewClient(
		option.WithBaseURL(server.URL),
		option.WithHTTPClient(server.Client()),
		option.WithAPIKey("dummy"),
	)
	purger := NewCloudflareCachePurger(client, "test", "https://api/categories", "https://api/products", "https://api/menu")
	if err := purger.Categories(context.Background()); err != nil {
		t.Fatalf("Failed to purge categories cache: %v", err)
	}
}

func TestCloudflareCachePurgerProducts(t *testing.T) {
	server := newTestServer()
	t.Cleanup(server.Close)

	client := cloudflare.NewClient(
		option.WithBaseURL(server.URL),
		option.WithHTTPClient(server.Client()),
		option.WithAPIKey("dummy"),
	)
	purger := NewCloudflareCachePurger(client, "test", "https://api/categories", "https://api/products", "https://api/menu")
	if err := purger.Products(context.Background()); err != nil {
		t.Fatalf("Failed to purge products cache: %v", err)
	}
}
