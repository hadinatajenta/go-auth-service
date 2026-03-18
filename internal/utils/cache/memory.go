package cache

import (
	"context"
	"errors"
	"sync"
	"time"
)

type item struct {
	value      interface{}
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

	s, ok := it.value.(string)
	if !ok {
		return "", errors.New("value is not a string")
	}

	return s, nil
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

func (c *memoryCache) SAdd(ctx context.Context, key string, members ...string) error {
	val, _ := c.items.LoadOrStore(key, item{
		value:      make(map[string]struct{}),
		expiration: 0, // Sets in this context are permanently cached until invalidated
	})

	it := val.(item)
	set, ok := it.value.(map[string]struct{})
	if !ok {
		return errors.New("key exists but is not a set")
	}

	for _, m := range members {
		set[m] = struct{}{}
	}

	return nil
}

func (c *memoryCache) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	val, ok := c.items.Load(key)
	if !ok {
		return false, errors.New("key not found")
	}

	it := val.(item)
	set, ok := it.value.(map[string]struct{})
	if !ok {
		return false, errors.New("value is not a set")
	}

	_, exists := set[member]
	return exists, nil
}

func (c *memoryCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	c.items.Range(func(key, value interface{}) bool {
		k, ok := key.(string)
		if ok && (prefix == "*" || (len(k) >= len(prefix) && k[:len(prefix)] == prefix)) {
			c.items.Delete(key)
		}
		return true
	})
	return nil
}
