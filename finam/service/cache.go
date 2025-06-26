package service

import "context"

type CacheRepo interface {
	Set(ctx context.Context, values ...string)
	GetLatest(ctx context.Context) (string, error)
	GetNLatest(ctx context.Context, n int) ([]string, error)
}

type Cache struct {
	cacheRepo CacheRepo
}

func NewCache(cacheRepo CacheRepo) *Cache {
	return &Cache{cacheRepo: cacheRepo}
}

func (c *Cache) Set(ctx context.Context, values ...string) {
	c.cacheRepo.Set(ctx, values...)
}

func (c *Cache) GetLatest(ctx context.Context) (string, error) {
	return c.GetLatest(ctx)
}

func (c *Cache) GetNLatest(ctx context.Context, n int) ([]string, error) {
	return c.cacheRepo.GetNLatest(ctx, n)
}
