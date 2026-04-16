package websocket

import "encoding/json"

// Event represents a valid websocket event.
type Event string

const (
	EventError Event = "error"

	EventAddSession   Event = "add_session"
	EventSessionAdded Event = "session_added"

	EventUpdateSession  Event = "update_session"
	EventSessionUpdated Event = "session_updated"

	EventAddOrderedProduct   Event = "add_ordered_product"
	EventOrderedProductAdded Event = "ordered_product_added"
)

// Message represent a websocket message.
type Message struct {
	Event Event           `json:"event"`
	Data  json.RawMessage `json:"data"`
}
