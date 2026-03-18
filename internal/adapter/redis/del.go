package redis

import (
	"context"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
)

func (r *Redis) Del(
	ctx context.Context,
	key string,
) error {
	removed, err := r.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	if removed == 0 {
		return domain.ErrKeyNotFound
	}

	return nil
}
