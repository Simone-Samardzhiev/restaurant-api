package testutils

import "testing"

// CheckEntities is a test helper that check entities that are identifiable by id.
// It checks if all ids match the ids of the entities by calling getId.
func CheckEntities[T any, K comparable](t testing.TB, expectedIds []K, entities []T, getId func(T) K) {
	t.Helper()

	if len(expectedIds) != len(entities) {
		t.Errorf("want %d entities, got %d", len(expectedIds), len(entities))
	}

	counter := make(map[K]int, len(expectedIds))
	for _, id := range expectedIds {
		counter[id]++
	}

	for _, entity := range entities {
		if counter[getId(entity)] == 0 {
			t.Errorf("missing entity with id: %v", getId(entity))
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
