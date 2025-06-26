package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{rdb: rdb}
}

func (c *Cache) Set(ctx context.Context, values ...string) {
	c.rdb.RPush(ctx, "balance_list", values)
}

func (c *Cache) GetLatest(ctx context.Context) (string, error) {
	vals, err := c.rdb.LRange(ctx, "balance_list", -1, -1).Result()
	if err != nil {
		return "", err
	}
	if len(vals) < 1 {
		return "", redis.Nil
	}
	return vals[0], nil
}

func (c *Cache) GetNLatest(ctx context.Context, n int) ([]string, error) {
	ln, err := c.rdb.LLen(ctx, "balance_list").Result()
	if err != nil {
		return []string{}, err
	}
	if ln >= int64(n) {
		vals, err := c.rdb.LRange(ctx, "balance_list", int64(-n), -1).Result()
		return vals, err
	}
	if ln == 0 {
		return []string{}, redis.Nil
	}
	vals, err := c.rdb.LRange(ctx, "balance_list", -ln, -1).Result()
	return vals, err
}
