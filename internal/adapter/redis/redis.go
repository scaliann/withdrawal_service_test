package redis

import "github.com/scaliann/withdrawal_service_test/pkg/redis"

type Redis struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{
		client: client,
	}
}
