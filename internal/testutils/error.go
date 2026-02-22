package testutils

import (
	"errors"
	"restaurant/internal/domain"
	"testing"
)

// AssertError is helper function for asserting an error is of type [domain.Error]
// and [domain.Error.Kind] and [domain.Error.Code] matches with the arguments.
func AssertError(t *testing.T, err error, kind domain.ErrorKind, code domain.ErrorCode, expectedDetailCodes ...domain.ErrorCode) {
	t.Helper()

	domainErr, ok := errors.AsType[*domain.Error](err)
	if !ok {
		t.Fatalf("want error type %T, got %T", (*domain.Error)(nil), err)
	}

	if domainErr.Kind != kind {
		t.Errorf("want error kind %s, got %s", kind, domainErr.Kind)
	}

	if domainErr.Code != code {
		t.Errorf("want error code %d, got %d", code, domainErr.Code)
	}

	if len(expectedDetailCodes) > 0 {
		checkDetailsCodes(t, expectedDetailCodes, domainErr.Details)
	}
}

// checkDetailsCodes checks if error codes in match the error codes in the details, in any order.
func checkDetailsCodes(t *testing.T, expectedCodes []domain.ErrorCode, details []domain.ErrorDetail) {
	t.Helper()

	if len(expectedCodes) != len(details) {
		t.Errorf("want %d codes, got %d", len(expectedCodes), len(details))
	}

	counter := make(map[domain.ErrorCode]int)
	for _, code := range expectedCodes {
		counter[code]++
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

// AssertErrorDetail is helper function for asserting an error is of type [domain.ErrorDetail]
// and [domain.ErrorDetail.Code] matches with the argument.
func AssertErrorDetail(t *testing.T, err error, code domain.ErrorCode) {
	t.Helper()

	detailErr, ok := errors.AsType[*domain.ErrorDetail](err)
	if !ok {
		t.Fatalf("want error type %T, got %T", (*domain.ErrorDetail)(nil), err)
	}

	if detailErr.Code != code {
		t.Errorf("want error code %s, got %s", code, detailErr.Code)
	}
}
