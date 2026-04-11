package websocket

import "encoding/json"

// Event represents a valid websocket event.
type Event string

const (
	ErrorEvent Event = "error"

	AddSessionEvent   Event = "add_session"
	SessionAddedEvent Event = "session_added"

	UpdateSessionEvent  Event = "update_session"
	SessionUpdatedEvent Event = "session_updated"
)

// Message represent a websocket message.
type Message struct {
	Event Event           `json:"event"`
	Data  json.RawMessage `json:"data"`
}
