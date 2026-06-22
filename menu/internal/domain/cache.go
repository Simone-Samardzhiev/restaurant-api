package domain

import "context"

// CachePurger describes how cache is purged.
type CachePurger interface {
	// Categories purges the cache for the categories.
	Categories(ctx context.Context) error

	// Products purges the cache for products.
	Products(ctx context.Context) error
}
