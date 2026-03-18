package redis

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"github.com/scaliann/withdrawal_service_test/internal/domain"
)

func (r *Redis) Get(
	ctx context.Context,
	key string,
) ([]byte, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrKeyNotFound
		}
		return nil, err
	}
	return data, nil
}
