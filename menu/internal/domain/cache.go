package domain

import "context"

// CachePurger describes how cache is purged.
type CachePurger interface {
	// Menu purges the cache for the menu.
	Menu(ctx context.Context) error
}
