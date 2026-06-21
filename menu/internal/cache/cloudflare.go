package cache

import (
	"context"
	"menu/internal/domain"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/cache"
)

type CloudflareCachePurger struct {
	client     *cloudflare.Client
	zoneId     string
	menuPrefix string
}

var _ domain.CachePurger = (*CloudflareCachePurger)(nil)

func NewCloudflareCachePurger(client *cloudflare.Client, zoneId, menuPrefix string) *CloudflareCachePurger {
	return &CloudflareCachePurger{
		client:     client,
		zoneId:     zoneId,
		menuPrefix: menuPrefix,
	}
}

func (c *CloudflareCachePurger) Menu(ctx context.Context) error {
	_, err := c.client.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(c.zoneId),
		Body: cache.CachePurgeParamsBodyCachePurgeFlexPurgeByPrefixes{
			Prefixes: cloudflare.F([]string{c.menuPrefix}),
		},
	})
	if err != nil {
		return domain.NewError("error purging menu cache", domain.ErrorCodeInternal, err)
	}
	return nil
}
