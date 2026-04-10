package order

import (
	"restaurant/internal/domain"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
)

func TestParseSessionTable(t *testing.T) {
	tests := []struct {
		name          string
		tableNumber   int
		wantErr       bool
		wantErrorCode domain.ErrorCode
	}{
		{
			name:        "valid",
			tableNumber: 1,
		},
		{
			name:          "invalid negative",
			tableNumber:   -1,
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidSessionTable,
		},
		{
			name:          "invalid table zero",
			tableNumber:   0,
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidSessionTable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseSessionTable(tt.tableNumber)
			if tt.wantErr {
				test.AssertErrorDetail(t, err, tt.wantErrorCode)
				return
			}

			if err != nil {
				t.Errorf("want no error , got %v", err)
			}
		})
	}
}

func TestParseSessionStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		wantErr       bool
		wantErrorCode domain.ErrorCode
	}{
		{
			name:   "valid",
			status: "paid",
		},
		{
			name:          "invalid",
			status:        "invalid",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidSessionStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseSessionStatus(tt.status)
			if tt.wantErr {
				test.AssertErrorDetail(t, err, tt.wantErrorCode)
				return
			}

			if err != nil {
				t.Errorf("want no error , got %v", err)
			}
		})
	}
}

func TestParseSession(t *testing.T) {
	tests := []struct {
		name             string
		status           string
		tableNumber      int
		wantErr          bool
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:        "valid",
			status:      "open",
			tableNumber: 1,
		},
		{
			name:             "invalid table number",
			status:           "open",
			tableNumber:      -1,
			wantErr:          true,
			wantErrorCode:    domain.ErrorCodeInvalidSession,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeInvalidSessionTable},
		},
		{
			name:             "invalid status",
			status:           "invalid",
			tableNumber:      1,
			wantErr:          true,
			wantErrorCode:    domain.ErrorCodeInvalidSession,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeInvalidSessionStatus},
		},
		{
			name:             "invalid table number and status",
			tableNumber:      -1,
			status:           "invalid",
			wantErr:          true,
			wantErrorCode:    domain.ErrorCodeInvalidSession,
			wantDetailsCodes: []domain.ErrorCode{domain.ErrorCodeInvalidSessionStatus, domain.ErrorCodeInvalidSessionTable},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseSession(uuid.New(), tt.tableNumber, tt.status)
			if tt.wantErr {
				test.AssertError(t, err, domain.ErrorKindValidation, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Errorf("want no error got %v", err)
			}
		})
	}
}
