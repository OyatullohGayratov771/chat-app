package redis

import (
	"context"
	"user-service/internal/config"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(cfg config.Config) *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: "",
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		panic("redis ulanmadi: " + err.Error())
	}
	return client
}
