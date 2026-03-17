package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

type item struct {
	value      string
	expiration int64
}

type memoryCache struct {
	items sync.Map
}

func NewMemoryCache() Cache {
	return &memoryCache{}
}

func (c *memoryCache) Get(ctx context.Context, key string) (string, error) {
	val, ok := c.items.Load(key)
	if !ok {
		return "", errors.New("key not found")
	}

	it := val.(item)
	if it.expiration > 0 && it.expiration < time.Now().UnixNano() {
		c.items.Delete(key)
		return "", errors.New("key expired")
	}

	return it.value, nil
}

func (c *memoryCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.items.Store(key, item{
		value:      value,
		expiration: exp,
	})
	return nil
}

func (c *memoryCache) Delete(ctx context.Context, key string) error {
	c.items.Delete(key)
	return nil
}
