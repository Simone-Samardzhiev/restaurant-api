package rest

import (
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func TestAddProductRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   *AddProductRequest
		wantField string
	}{
		{
			name: "valid",
			request: &AddProductRequest{
				Name:             "Valid name",
				Description:      "Valid description",
				Price:            decimal.NewFromFloat(1.00),
				CategoryID:       uuid.New(),
				ImageContentType: "image/png",
			},
		},
		{
			name: "invalid name",
			request: &AddProductRequest{
				Name:             "",
				Description:      "Valid description",
				Price:            decimal.NewFromFloat(1.00),
				CategoryID:       uuid.New(),
				ImageContentType: "image/png",
			},
			wantField: "name",
		},
		{
			name: "invalid description",
			request: &AddProductRequest{
				Name:             "Valid name",
				Description:      "",
				Price:            decimal.NewFromFloat(1.00),
				CategoryID:       uuid.New(),
				ImageContentType: "image/png",
			},
			wantField: "description",
		},
		{
			name: "invalid price",
			request: &AddProductRequest{
				Name:        "Valid name",
				Description: "Valid description",
				Price:       decimal.NewFromFloat(-10),
				CategoryID:  uuid.New(),
			},
			wantField: "price",
		},
		{
			name: "invalid image content type",
			request: &AddProductRequest{
				Name:             "Valid name",
				Description:      "Valid description",
				Price:            decimal.NewFromFloat(1.00),
				CategoryID:       uuid.New(),
				ImageContentType: "image/pn",
			},
			wantField: "imageContentType",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.request.Validate()
			if tt.wantField == "" {
				if got != nil {
					t.Errorf("Validate() = %v, want nil", got)
				}

				return
			}

			_, ok := got[tt.wantField]
			if !ok {
				t.Errorf("Missing field %v", tt.wantField)
			}
		})
	}
}
