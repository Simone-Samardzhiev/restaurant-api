package rest_test

import (
	"restaurant/internal/adapter/handler/rest/middleware"
	"restaurant/internal/domain"
	"testing"
)

// matchErrorCodes verifies details error codes match the codes in details
// ignoring the order.
func matchErrorCodes(t *testing.T, expectedCodes []domain.ErrorCode, details []middleware.ErrorDetailsResponse) {
	t.Helper()

	if len(details) != len(details) {
		t.Fatalf("length mismatch: expected codes %d, got %d ", len(expectedCodes), len(details))
	}

	var counter = map[string]int{}
	for _, code := range expectedCodes {
		counter[code.String()]++
	}

	for _, detail := range details {
		if counter[detail.Code] == 0 {
			t.Errorf("unexpected error code %s", detail.Code)
			continue
		}
		counter[detail.Code]--
	}

	for code, count := range counter {
		if count != 0 {
			t.Errorf("missing error code %s", code)
		}
	}
}
