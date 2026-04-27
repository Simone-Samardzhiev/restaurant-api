package rest

import (
	"testing"
)

func TestAddCategoryRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   *AddCategoryRequest
		wantField string
	}{
		{
			name: "valid",
			request: &AddCategoryRequest{
				Name: "test",
			},
		},
		{
			name: "short name",
			request: &AddCategoryRequest{
				Name: "",
			},
			wantField: "name",
		},
		{
			name: "long name",
			request: &AddCategoryRequest{
				Name: "testtesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttesttest",
			},
			wantField: "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.request.Validate()

			if tt.wantField == "" {
				if got != nil {
					t.Errorf("Validate() = %v, want nil", got)
				}
				return
			}

			_, ok := got[tt.wantField]
			if !ok {
				t.Errorf("Missing field: %s", tt.wantField)
			}
		})
	}

}
