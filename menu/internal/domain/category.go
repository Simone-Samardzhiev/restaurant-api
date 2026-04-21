package domain

import (
	"time"

	"github.com/google/uuid"
)

// Category represents a category of products in the menu.
type Category struct {
	Id        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
