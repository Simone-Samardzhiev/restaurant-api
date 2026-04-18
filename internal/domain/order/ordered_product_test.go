package order_test

import (
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
)

func TestParseOrderedProductStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		wantErr       bool
		wantErrorCode domain.ErrorCode
	}{
		{
			name:   "valid",
			status: "pending",
		},
		{
			name:          "invalid",
			status:        "invalid",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidOrderedProductStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := order.ParseOrderedProductStatus(tt.status)
			if tt.wantErr {
				test.AssertErrorDetail(t, err, tt.wantErrorCode)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}
		})
	}
}

func TestParseOrderedProduct(t *testing.T) {
	tests := []struct {
		name             string
		status           string
		wantErr          bool
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:   "valid",
			status: "pending",
		},
		{
			name:             "invalid",
			status:           "invalid",
			wantErr:          true,
			wantErrorCode:    domain.ErrorCodeInvalidOrderedProduct,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeInvalidOrderedProductStatus},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := order.ParseOrderedProduct(uuid.Nil, uuid.Nil, uuid.Nil, tt.status)
			if tt.wantErr {
				test.AssertError(t, err, domain.ErrorKindValidation, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}
		})
	}
}
