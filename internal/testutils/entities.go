package testutils

import "testing"

// CheckEntities is a test helper that check entities that are identifiable by id.
// It checks if the ids match the ids of the entities by calling getId.
func CheckEntities[T any, K comparable](t testing.TB, wantIds []K, entities []T, getId func(T) K) {
	t.Helper()

	if len(wantIds) != len(entities) {
		t.Errorf("want %d ids, got %d", len(wantIds), len(entities))
	}

	counter := make(map[K]int, len(wantIds))
	for _, id := range wantIds {
		counter[id]++
	}

	for _, entity := range entities {
		if counter[getId(entity)] == 0 {
			t.Errorf("unexpected : %v", getId(entity))
			continue
		}
		counter[getId(entity)]--
	}

	for id, count := range counter {
		if count != 0 {
			t.Errorf("missing entity with id: %v", id)
		}
	}
}
