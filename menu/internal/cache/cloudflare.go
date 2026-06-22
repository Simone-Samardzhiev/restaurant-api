package cache

import (
	"context"
	"menu/internal/domain"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/cache"
)

type CloudflareCachePurger struct {
	client        *cloudflare.Client
	zoneId        string
	categoriesUrl string
	productsUrl   string
	menuUrl       string
}

var _ domain.CachePurger = (*CloudflareCachePurger)(nil)

func NewCloudflareCachePurger(client *cloudflare.Client, zoneId, categoriesUrl, productsUrl, menuUrl string) *CloudflareCachePurger {
	return &CloudflareCachePurger{
		client:        client,
		zoneId:        zoneId,
		categoriesUrl: categoriesUrl,
		productsUrl:   productsUrl,
		menuUrl:       menuUrl,
	}
}

func (c *CloudflareCachePurger) Categories(ctx context.Context) error {
	_, err := c.client.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(c.zoneId),
		Body: cache.CachePurgeParamsBodyCachePurgeSingleFile{
			Files: cloudflare.F([]string{c.categoriesUrl, c.menuUrl}),
		},
	})
	if err != nil {
		return domain.NewError("error purging menu cache", domain.ErrorCodeInternal, err)
	}
	return nil
}

func (c *CloudflareCachePurger) Products(ctx context.Context) error {
	_, err := c.client.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(c.zoneId),
		Body: cache.CachePurgeParamsBodyCachePurgeSingleFile{
			Files: cloudflare.F([]string{c.productsUrl, c.menuUrl}),
		},
	})
	if err != nil {
		return domain.NewError("error purging menu cache", domain.ErrorCodeInternal, err)
	}
	return nil
}
