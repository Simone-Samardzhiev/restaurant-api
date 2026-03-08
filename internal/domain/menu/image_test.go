package menu_test

import (
	"restaurant/internal/domain"
	"restaurant/internal/domain/menu"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseImageType(tt.imageType)
			if tt.wantErr {
				test.AssertErrorDetail(t, err, tt.wantErrorCode)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}
			if parsed.String() != tt.imageType {
				t.Fatalf("want image type %s, got %s", tt.imageType, parsed.String())
			}
		})
	}
}

func TestParseUpdateImageRequest(t *testing.T) {
	tests := []struct {
		name             string
		imageType        string
		wantErr          bool
		wantErrorCode    domain.ErrorCode
		wantDetailsCodes []domain.ErrorCode
	}{
		{
			name:      "valid image",
			imageType: "png",
		},
		{
			name:          "invalid image",
			imageType:     "invalid image",
			wantErr:       true,
			wantErrorCode: domain.ErrorCodeInvalidImageUpdate,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeInvalidImageType,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parsed, err := menu.ParseUpdateImageRequest(uuid.New(), nil, tt.imageType)
			if tt.wantErr {
				test.AssertError(t, err, domain.ErrorKindValidation, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}
			if err != nil {
				t.Fatalf("want no error, got %v", err)
			}

			if tt.imageType != parsed.ImageType.String() {
				t.Fatalf("want image type %s, got %s", tt.imageType, parsed.ImageType.String())
			}
		})
	}
}
