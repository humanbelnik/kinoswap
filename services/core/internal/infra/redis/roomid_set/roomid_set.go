package infra_redis_roomid_set

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/humanbelnik/kinoswap/core/internal/model"
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

func (d *Driver) Remove(ctx context.Context) (model.RoomID, error) {
	id, err := d.client.SPop(d.key).Result()
	if err == redis.Nil {
		return model.EmptyRoomID, nil
	}
	if err != nil {
		return model.EmptyRoomID, err
	}
	return model.RoomID(id), nil
}

func (d *Driver) Add(ctx context.Context, roomID model.RoomID) error {
	if roomID == model.EmptyRoomID {
		return nil
	}

	if err := d.client.SAdd(d.key, string(roomID)).Err(); err != nil {
		return err
	}
	return nil
}
