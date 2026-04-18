package order_test

import (
	"context"
	"restaurant/internal/domain"
	"restaurant/internal/domain/order"
	"restaurant/internal/test"
	"testing"

	"github.com/google/uuid"
)

type fakeSessionRepository struct {
	onSave    func(ctx context.Context, request *order.AddSessionRequest) (*order.Session, error)
	onGet     func(ctx context.Context) ([]order.Session, error)
	onGetById func(ctx context.Context, id uuid.UUID) (*order.Session, error)
	onUpdate  func(ctx context.Context, request *order.UpdateSessionRequest) error
}

var _ order.SessionRepository = (*fakeSessionRepository)(nil)

func (r *fakeSessionRepository) Save(ctx context.Context, request *order.AddSessionRequest) (*order.Session, error) {
	if r.onSave != nil {
		panic("onSave not implemented")
	}
	return r.onSave(ctx, request)
}

func (r *fakeSessionRepository) Get(ctx context.Context) ([]order.Session, error) {
	if r.onGet != nil {
		panic("onGet not implemented")
	}
	return r.onGet(ctx)
}

func (r *fakeSessionRepository) GetById(ctx context.Context, id uuid.UUID) (*order.Session, error) {
	if r.onGetById != nil {
		panic("onGetById not implemented")
	}
	return r.onGetById(ctx, id)
}

func (r *fakeSessionRepository) Update(ctx context.Context, request *order.UpdateSessionRequest) error {
	if r.onUpdate != nil {
		panic("onUpdate not implemented")
	}
	return r.onUpdate(ctx, request)
}

type fakeOrderedProductRepository struct {
	onSave           func(ctx context.Context, request *order.AddOrderedProductRequest) (*order.OrderedProduct, error)
	onGetBySessionId func(ctx context.Context, sessionId uuid.UUID) ([]order.OrderedProduct, error)
}

func (r *fakeOrderedProductRepository) Save(ctx context.Context, request *order.AddOrderedProductRequest) (*order.OrderedProduct, error) {
	if r.onSave != nil {
		panic("onSave not implemented")
	}
	return r.onSave(ctx, request)
}

func (r *fakeOrderedProductRepository) GetBySessionId(ctx context.Context, sessionId uuid.UUID) ([]order.OrderedProduct, error) {
	if r.onGetBySessionId != nil {
		panic("onGetBySessionId not implemented")
	}
	return r.onGetBySessionId(ctx, sessionId)
}

var _ order.OrderedProductRepository = (*fakeOrderedProductRepository)(nil)

func TestDefaultSessionServiceGetSessionDetails(t *testing.T) {
	tests := []struct {
		name                     string
		id                       uuid.UUID
		sessionRepository        order.SessionRepository
		orderedProductRepository order.OrderedProductRepository
		wantDetails              *order.SessionDetails
		wantErr                  bool
		wantErrorKind            domain.ErrorKind
		wantErrorCode            domain.ErrorCode
		wantDetailsCodes         []domain.ErrorCode
	}{
		{
			name: "success",
			id:   uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			sessionRepository: &fakeSessionRepository{
				onGetById: func(ctx context.Context, id uuid.UUID) (*order.Session, error) {
					if id != uuid.MustParse("00000000-0000-0000-0000-000000000000") {
						panic("wrong id")
					}

					return test.Must(order.ParseSession(id, 10, "open")), nil
				},
			},
			orderedProductRepository: &fakeOrderedProductRepository{
				onGetBySessionId: func(ctx context.Context, sessionId uuid.UUID) ([]order.OrderedProduct, error) {
					if sessionId != uuid.MustParse("00000000-0000-0000-0000-000000000000") {
						panic("wrong id")
					}

					return []order.OrderedProduct{
						{
							Id:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
							ProductId: uuid.MustParse("00000000-0000-0000-0000-000000000002"),
							SessionId: sessionId,
							Status:    order.StatusPending,
						},
					}, nil
				},
			},
		},
		{
			name: "session not found",
			id:   uuid.New(),
			sessionRepository: &fakeSessionRepository{
				onGetById: func(ctx context.Context, id uuid.UUID) (*order.Session, error) {
					return nil, domain.NewNotFoundError(
						"session not found",
						domain.ErrorCodeSessionNotFound,
						domain.ErrorDetail{
							Code:    domain.ErrorCodeSessionNotFoundByID,
							Message: "session not found by id",
						})
				},
			},
			wantErr:       true,
			wantErrorKind: domain.ErrorKindNotFound,
			wantErrorCode: domain.ErrorCodeSessionNotFound,
			wantDetailsCodes: []domain.ErrorCode{
				domain.ErrorCodeSessionNotFoundByID,
			},
		},
		{
			id: uuid.MustParse("00000000-0000-0000-0000-000000000000"),
			sessionRepository: &fakeSessionRepository{
				onGetById: func(ctx context.Context, id uuid.UUID) (*order.Session, error) {
					if id != uuid.MustParse("00000000-0000-0000-0000-000000000000") {
						panic("wrong id")
					}

					return test.Must(order.ParseSession(id, 10, "open")), nil
				},
			},
			orderedProductRepository: &fakeOrderedProductRepository{
				onGetBySessionId: func(ctx context.Context, sessionId uuid.UUID) ([]order.OrderedProduct, error) {
					if sessionId != uuid.MustParse("00000000-0000-0000-0000-000000000000") {
						panic("wrong id")
					}

					return nil, domain.NewInternalError("error getting ordered products", nil)
				},
			},
			wantErr:       true,
			wantErrorKind: domain.ErrorKindInternal,
			wantErrorCode: domain.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			service := order.NewDefaultSessionService(tt.sessionRepository, tt.orderedProductRepository)

			_, err := service.GetSessionDetails(context.Background(), tt.id)
			if tt.wantErr {
				test.AssertError(t, err, tt.wantErrorKind, tt.wantErrorCode, tt.wantDetailsCodes...)
				return
			}

			if err != nil {
				t.Fatalf("want no error, got: %v", err)
			}
		})
	}

}
