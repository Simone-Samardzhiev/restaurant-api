package menu_test

import (
	"restaurant/internal/domain"
	"testing"
)

// matchErrorCodes verifies details error codes match the codes in details
// ignoring the order.
func matchErrorCodes(t *testing.T, expectedCodes []domain.ErrorCode, details []domain.ErrorDetail) {
	t.Helper()

	if len(details) != len(expectedCodes) {
		t.Fatalf("length mismatch: expected codes %d, got %d ", len(expectedCodes), len(details))
	}

	var counter = map[domain.ErrorCode]int{}
	for _, code := range expectedCodes {
		counter[code]++
	}

	for _, detail := range details {
		if counter[detail.Code] == 0 {
			t.Errorf("unexpected error code: %s", detail.Code)
		}
		counter[detail.Code]--
	}

	for code, count := range counter {
		if count != 0 {
			t.Errorf("missing error code: %s", code)
		}
	}
}
