package rate

import (
	"menu/internal/config"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
)

func TestValkeyStoreAllow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	url := os.Getenv("TEST_VALKEY_URL")
	valkeyOption, err := valkey.ParseURL(url)
	if err != nil {
		t.Fatalf("Error parsing valkey url: %v", err)
	}

	client, err := valkey.NewClient(valkeyOption)
	if err != nil {
		t.Fatalf("Error connecting to valkey: %v", err)
	}

	limit := 10
	store := NewValkeyStore(client, config.RateLimit{
		Limit:  limit,
		Window: time.Second,
	})

	id := uuid.NewString()

	for i := 0; i < limit; i++ {
		result, err := store.Allow(id)
		if err != nil {
			t.Fatalf("Error allowing id %s: %v", id, err)
		}

		if !result {
			t.Fatalf("Error allowing id %s: result should be true", id)
		}
	}

	result, err := store.Allow(id)
	if err != nil {
		t.Fatalf("Error allowing id %s: %v", id, err)
	}
	if result {
		t.Fatalf("Error allowing id %s: result should be false", id)
	}

	time.Sleep(time.Second + 500*time.Millisecond)
	result, err = store.Allow(id)
	if err != nil {
		t.Fatalf("Error allowing id %s: %v", id, err)
	}
	if !result {
		t.Fatalf("Error allowing id %s: result should be true", id)
	}
}
