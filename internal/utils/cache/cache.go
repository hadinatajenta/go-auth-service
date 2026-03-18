package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	
	// Set operations (for permissions)
	SAdd(ctx context.Context, key string, members ...string) error
	SIsMember(ctx context.Context, key string, member string) (bool, error)
	DeleteByPrefix(ctx context.Context, prefix string) error
}
