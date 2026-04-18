package rest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"restaurant/internal/adapter/handler/rest"
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"
	"restaurant/internal/test"
	"testing"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
)

type fakeSessionService struct {
	onAddSession        func(ctx context.Context, request *order.AddSessionRequest) (*order.Session, error)
	onUpdateSession     func(ctx context.Context, request *order.UpdateSessionRequest) error
	onGetSessionDetails func(ctx context.Context, id uuid.UUID) (*order.SessionDetails, error)
}

func (s *fakeSessionService) AddSession(ctx context.Context, request *order.AddSessionRequest) (*order.Session, error) {
	if s.onAddSession == nil {
		panic("onAddSession not implemented")
	}
	return s.onAddSession(ctx, request)
}

func (s *fakeSessionService) UpdateSession(ctx context.Context, request *order.UpdateSessionRequest) error {
	if s.onUpdateSession == nil {
		panic("onUpdateSession not implemented")
	}
	return s.onUpdateSession(ctx, request)
}

func (s *fakeSessionService) GetSessionDetails(ctx context.Context, id uuid.UUID) (*order.SessionDetails, error) {
	if s.onGetSessionDetails == nil {
		panic("onGetSessionDetails not implemented")
	}
	return s.onGetSessionDetails(ctx, id)
}

var _ order.SessionService = (*fakeSessionService)(nil)

func TestSessionHandlerGetSessionDetails(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		service        order.SessionService
		wantHttpStatus int
		wantErrorCode  domain.ErrorCode
	}{
		{
			name: "success",
			id:   uuid.New().String(),
			service: &fakeSessionService{
				onGetSessionDetails: func(ctx context.Context, id uuid.UUID) (*order.SessionDetails, error) {
					return &order.SessionDetails{
						Session: order.Session{
							Id:     id,
							Table:  test.Must(order.ParseSessionTable(10)),
							Status: order.StatusOpen,
						},
					}, nil
				},
			},
			wantHttpStatus: http.StatusOK,
		},
		{
			name:           "invalid uuid",
			id:             "invalid",
			service:        &fakeSessionService{},
			wantHttpStatus: http.StatusBadRequest,
			wantErrorCode:  domain.ErrorCodeInvalidUUID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			router := test.NewSessionRouter(tt.service)
			request := httptest.NewRequest(http.MethodGet, "/sessions/"+tt.id, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantHttpStatus {
				t.Fatalf("want status %d, got %d", tt.wantHttpStatus, recorder.Code)
			}

			if recorder.Code == http.StatusOK {
				var response rest.SessionDetailsResponse
				err := json.NewDecoder(recorder.Body).Decode(&response)
				if err != nil {
					t.Fatalf("error decoding response body: %v", err)
				}
			} else {
				test.CheckErrorResponse(t, recorder.Body, tt.wantErrorCode)
			}
		})
	}
}
