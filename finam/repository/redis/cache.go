package redis

import (
	"context"
	"fmt"

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
	fmt.Println(len(vals))
	return vals[0], nil
}
