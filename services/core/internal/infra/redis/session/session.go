package infra_session_cache

import (
	"time"

	"github.com/go-redis/redis"
)

type Driver struct {
	client *redis.Client
	key    string
}

func New(
	client *redis.Client,
	key string,
) *Driver {
	return &Driver{
		client: client,
		key:    key,
	}
}

func (d *Driver) Set(key string, value string, ttl time.Duration) error {
	fullKey := d.getFullKey(key)
	err := d.client.Set(fullKey, value, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}
func (d *Driver) Get(key string) (string, error) {
	fullKey := d.getFullKey(key)

	val, err := d.client.Get(fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}
		return "", err
	}

	return val, nil
}

func (d *Driver) getFullKey(key string) string {
	if d.key != "" {
		return d.key + ":" + key
	}
	return key
}
