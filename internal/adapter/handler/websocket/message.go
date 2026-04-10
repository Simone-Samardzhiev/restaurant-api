package websocket

import "encoding/json"

// Event represents a valid websocket event.
type Event string

const (
	AddSessionEvent   Event = "add_session"
	SessionAddedEvent Event = "session_added"
	ErrorEvent        Event = "error"
)

// Message represent a websocket message.
type Message struct {
	Event Event           `json:"event"`
	Data  json.RawMessage `json:"data"`
}
