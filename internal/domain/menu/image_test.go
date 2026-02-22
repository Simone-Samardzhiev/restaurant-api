package menu_test

import (
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/testutils"
	"testing"
)

func TestParseImageType(t *testing.T) {
	tests := []struct {
		name          string
		imageType     string
		wantErr       bool
		wantErrorCode domain.ErrorCode
	}{
		{
			name:      "valid image1",
			imageType: "png",
		},
		{
			name:      "valid image2",
			imageType: "jpg",
		},
		{
			name:      "valid image3",
			imageType: "jpeg",
		},
		{
			name:          "invalid image",
			imageType:     "invalid image",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidImageType,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseImageType(test.imageType)
			if test.wantErr {
				testutils.AssertErrorDetail(t, err, test.wantErrorCode)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}
			if parsed.String() != test.imageType {
				t.Fatalf("want image type %s, got %s", test.imageType, parsed.String())
			}
		})
	}
}
