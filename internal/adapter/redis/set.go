package redis

import (
	"context"
	"time"
)

func (r *Redis) Set(
	ctx context.Context,
	key string,
	value []byte,
	ttl time.Duration,
) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}
