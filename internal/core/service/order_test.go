package service_test

import (
	"context"
	"restaurant/internal/core/domain"
	"restaurant/internal/core/port/mock"
	"restaurant/internal/core/service"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestOrderService_UpdateSession(t *testing.T) {
	tests := []struct {
		name          string
		expectedError error
		update        *domain.UpdateOrderSessionDTO
		mockSetup     func(orderRepository *mock.MockOrderRepository)
	}{
		{
			name:   "success",
			update: domain.NewUpdateOrderSessionDTO(uuid.Nil, new(int), new(domain.OrderSessionStatus)),
			mockSetup: func(orderRepository *mock.MockOrderRepository) {
				orderRepository.EXPECT().
					UpdateSession(
						gomock.AssignableToTypeOf(context.Background()),
						gomock.AssignableToTypeOf(&domain.UpdateOrderSessionDTO{}),
					).Return(nil)
			},
		},
		{
			name:          "nothing to update",
			update:        domain.NewUpdateOrderSessionDTO(uuid.Nil, nil, nil),
			expectedError: domain.ErrNothingToUpdate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			orderRepository := mock.NewMockOrderRepository(ctrl)
			if tt.mockSetup != nil {
				tt.mockSetup(orderRepository)
			}

			err := service.NewOrderService(orderRepository).UpdateSession(context.Background(), tt.update)
			require.ErrorIs(t, err, tt.expectedError)
		})
	}
}
