package app

import (
	"github.com/humanbelnik/kinoswap/core/internal/config"
	http_init "github.com/humanbelnik/kinoswap/core/internal/delivery/http/init"
	http_room "github.com/humanbelnik/kinoswap/core/internal/delivery/http/room"
	infra_pg_init "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/init"
	infra_postgres_room "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/room"
	infra_redis_init "github.com/humanbelnik/kinoswap/core/internal/infra/redis/init"
	infra_redis_roomid_set "github.com/humanbelnik/kinoswap/core/internal/infra/redis/roomid_set"
	storage_room "github.com/humanbelnik/kinoswap/core/internal/storage/room"
	usecase_room "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
)

func Go(cfg *config.Config) {
	const (
		roomidSetKey = "room_id"
	)

	redisConn := infra_redis_init.MustEstablishConn(cfg.Redis)
	pgConn := infra_pg_init.MustEstablishConn(cfg.Postgres)

	roomIDSet := infra_redis_roomid_set.New(redisConn, roomidSetKey)
	roomDB := infra_postgres_room.New(pgConn)
	roomStorage := storage_room.New(roomDB, roomIDSet)
	roomUC := usecase_room.New(roomStorage)

	controllerPool := http_init.NewControllerPool()
	controllerPool.Add(http_room.New(roomUC))

	controllerPool.Register()
	controllerPool.RunAll(cfg.HTTP.Port)
}
