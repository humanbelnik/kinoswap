package infra_redis_init

import (
	"fmt"
	"log"

	"github.com/go-redis/redis"
	"github.com/humanbelnik/kinoswap/core/internal/config"
)

func MustEstablishConn(cfg config.RedisCache) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})

	if err := client.Ping().Err(); err != nil {
		log.Fatal("redis ping failed", err)
	}

	return client
}
