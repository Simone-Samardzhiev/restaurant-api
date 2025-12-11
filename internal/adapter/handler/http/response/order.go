package response

import (
	"restaurant/internal/core/domain"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// OrderSessionResponse represents an order response.
type OrderSessionResponse struct {
	Id          uuid.UUID                 `json:"id"`
	TableNumber int                       `json:"tableNumber"`
	Status      domain.OrderSessionStatus `json:"status"`
}

// NewOrderSessionResponse creates a new OrderSessionResponse instance.
func NewOrderSessionResponse(order *domain.OrderSession) OrderSessionResponse {
	return OrderSessionResponse{
		Id:          order.Id,
		TableNumber: order.TableNumber,
		Status:      order.Status,
	}
}

type BillItemResponse struct {
	Product    ProductResponse `json:"product"`
	Quantity   int             `json:"quantity"`
	TotalPrice decimal.Decimal `json:"totalPrice"`
}

func NewBillItemResponse(items []domain.BillItem) []BillItemResponse {
	response := make([]BillItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, BillItemResponse{
			Product: ProductResponse{
				Id:          item.Product.Id,
				Name:        item.Product.Name,
				Description: item.Product.Description,
				ImageUrl:    item.Product.ImageUrl,
				Category:    item.Product.Category,
				Price:       item.Product.Price,
			},
			Quantity:   item.Quantity,
			TotalPrice: item.TotalPrice,
		})
	}

	return response
}

// BillResponse represent a bill response.
type BillResponse struct {
	Products   []BillItemResponse `json:"products"`
	TotalPrice decimal.Decimal    `json:"totalPrice"`
}

// NewBillResponse creates a new BillResponse instance.
func NewBillResponse(bill *domain.Bill) *BillResponse {
	return &BillResponse{
		Products:   NewBillItemResponse(bill.Items),
		TotalPrice: bill.FullPrice,
	}
}
