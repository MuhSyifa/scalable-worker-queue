package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type DistributedLock struct {
	client *redis.Client
}

func NewDistributedLock(client *redis.Client) *DistributedLock {
	return &DistributedLock{client: client}
}

// Acquire tries to acquire a lock for a specific key
func (l *DistributedLock) Acquire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	// Use SetNX to only set the key if it doesn't already exist
	return l.client.SetNX(ctx, "lock:"+key, "1", expiration).Result()
}

// Release removes the lock for a specific key
func (l *DistributedLock) Release(ctx context.Context, key string) error {
	return l.client.Del(ctx, "lock:"+key).Err()
}
