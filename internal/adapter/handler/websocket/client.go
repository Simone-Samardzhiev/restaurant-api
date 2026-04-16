package websocket

import (
	"encoding/json"
	"restaurant/internal/domain/order"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

// Client represents a connected client.
type Client struct {
	Id        uuid.UUID
	SessionId uuid.UUID
	conn      *websocket.Conn

	orderedProductService order.OrderedProductService
	hub                   *Hub

	send chan []byte
}

// NewClient allocates and creates a new client with 256 buffered channel.
func NewClient(
	id, sessionId uuid.UUID,
	conn *websocket.Conn,
	orderedProductService order.OrderedProductService,
	hub *Hub,
) *Client {
	return &Client{
		Id:        id,
		SessionId: sessionId,
		conn:      conn,

		orderedProductService: orderedProductService,
		hub:                   hub,

		send: make(chan []byte, 256),
	}
}

// WritePump listens for messages over the send channel and sends them over the
// websocket connection.
func (c *Client) WritePump() {
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			return
		}
	}
}

// AddOrderedProductRequest represents the request data for adding a new ordered product.
type AddOrderedProductRequest struct {
	ProductId uuid.UUID `json:"productId"`
}

// OrderedProductResponse represent the response data for adding a new ordered product.
type OrderedProductResponse struct {
	Id        uuid.UUID `json:"id"`
	ProductId uuid.UUID `json:"productId"`
	Status    string    `json:"status"`
}

// addOrderedProduct handles [EventAddOrderedProduct].
func (c *Client) addOrderedProduct(message *Message) {
	var req AddOrderedProductRequest
	if err := json.Unmarshal(message.Data, &req); err != nil {
		c.send <- handleInvalidJSON(err)
		return
	}

	result, err := c.orderedProductService.PlaceOrder(context.Background(), order.NewAddOrderedProductRequest(req.ProductId, c.SessionId))
	if err != nil {
		c.send <- handleDomainError(err)
		return
	}

	data, err := json.Marshal(OrderedProductResponse{
		Id:        c.Id,
		ProductId: req.ProductId,
		Status:    result.Status.String(),
	})
	if err != nil {
		zap.L().Error("error encoding data", zap.Error(err))
	}

	response := Message{
		Event: EventOrderedProductAdded,
		Data:  data,
	}

	body, err := json.Marshal(response)
	if err != nil {
		zap.L().Error("error encoding body", zap.Error(err))
	}

	c.hub.Broadcast(c.SessionId, body)
}

// ReadPump reads events from websocket connection.
func (c *Client) ReadPump() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var message Message

		if err = json.Unmarshal(msg, &message); err != nil {
			c.send <- handleInvalidJSON(err)
			continue
		}

		switch message.Event {
		case EventAddOrderedProduct:
			c.addOrderedProduct(&message)
		default:
			c.send <- handleInvalidEvent()
		}
	}
}
