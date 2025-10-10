package app

import (
	"log/slog"

	"github.com/humanbelnik/kinoswap/core/internal/config"
	http_init "github.com/humanbelnik/kinoswap/core/internal/delivery/http/init"
	http_room "github.com/humanbelnik/kinoswap/core/internal/delivery/http/room"
	http_swagger "github.com/humanbelnik/kinoswap/core/internal/delivery/http/swagger"
	http_voting "github.com/humanbelnik/kinoswap/core/internal/delivery/http/voting"
	ws_room "github.com/humanbelnik/kinoswap/core/internal/delivery/ws/room"
	infra_embedder "github.com/humanbelnik/kinoswap/core/internal/infra/embedder"
	infra_postgres_embedding "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/embedding"
	infra_pg_init "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/init"
	infra_postgres_room "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/room"
	infra_postgres_vote "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/vote"
	infra_redis_init "github.com/humanbelnik/kinoswap/core/internal/infra/redis/init"
	infra_redis_roomid_set "github.com/humanbelnik/kinoswap/core/internal/infra/redis/roomid_set"
	usecase_room "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
	usecase_vote "github.com/humanbelnik/kinoswap/core/internal/usecase/vote"
)

func Go(cfg *config.Config) {
	const (
		roomIDSetKey = "room_id"
	)

	redisConn := infra_redis_init.MustEstablishConn(cfg.Redis)
	pgConn := infra_pg_init.MustEstablishConn(cfg.Postgres)
	embedderConn := infra_embedder.MustEstablishConnection(cfg.Embedder)

	roomIDSet := infra_redis_roomid_set.New(redisConn, roomIDSetKey)

	roomRepo := infra_postgres_room.New(pgConn)
	voteRepo := infra_postgres_vote.New(pgConn)
	//movieRepo := infra_postgres_meta.New(pgConn)
	embedderRepo := infra_postgres_embedding.New(pgConn)

	roomUC := usecase_room.New(roomRepo, embedderConn, embedderRepo, roomIDSet)
	voteUC := usecase_vote.New(voteRepo)
	//	movieUC := usecase_movie.New(embedder, movieRepo, embedderRepo)

	hub := ws_room.New(slog.Default())

	controllerPool := http_init.NewControllerPool()

	controllerPool.Add(http_swagger.New())
	controllerPool.Add(http_room.New(roomUC, hub))
	controllerPool.Add(http_voting.New(voteUC, hub))
	//controllerPool.Add(http_movie.New(movieUC, hub))

	controllerPool.Register()
	controllerPool.RunAll(cfg.HTTP.Port)
}
