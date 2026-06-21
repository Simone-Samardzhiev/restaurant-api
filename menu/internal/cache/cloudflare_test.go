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
		_, _ = w.Write([]byte(`{"success": true}`))
		w.WriteHeader(http.StatusOK)
	})

	return httptest.NewServer(mux)
}

func TestCloudflareCachePurgerPurgeMenuCache(t *testing.T) {
	server := newTestServer()
	t.Cleanup(server.Close)

	client := cloudflare.NewClient(
		option.WithBaseURL(server.URL),
		option.WithHTTPClient(server.Client()),
		option.WithAPIKey("dummy"),
	)
	purger := NewCloudflareCachePurger(client, "test", "api/v1/menu")

	if err := purger.Menu(context.Background()); err != nil {
		t.Fatalf("Failed to purge menu cache: %v", err)
	}
}
