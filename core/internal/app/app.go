package app

import (
	"log/slog"

	"github.com/humanbelnik/kinoswap/core/internal/config"
	http_init "github.com/humanbelnik/kinoswap/core/internal/delivery/http/init"
	http_room "github.com/humanbelnik/kinoswap/core/internal/delivery/http/room"
	http_voting "github.com/humanbelnik/kinoswap/core/internal/delivery/http/voting"
	http_movie "github.com/humanbelnik/kinoswap/core/internal/delivery/movie"
	ws_room "github.com/humanbelnik/kinoswap/core/internal/delivery/ws/room"
	infra_embedder_mock "github.com/humanbelnik/kinoswap/core/internal/infra/embedder"
	infra_pg_init "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/init"
	infra_postgres_meta "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/movie"
	infra_postgres_room "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/room"
	infra_postgres_vote "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/vote"
	infra_qdrant_mock "github.com/humanbelnik/kinoswap/core/internal/infra/qdrant"
	infra_redis_init "github.com/humanbelnik/kinoswap/core/internal/infra/redis/init"
	infra_redis_roomid_set "github.com/humanbelnik/kinoswap/core/internal/infra/redis/roomid_set"
	usecase_movie "github.com/humanbelnik/kinoswap/core/internal/usecase/movie"
	usecase_room "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
	usecase_vote "github.com/humanbelnik/kinoswap/core/internal/usecase/vote"
)

func Go(cfg *config.Config) {
	const (
		roomIDSetKey = "room_id"
	)

	redisConn := infra_redis_init.MustEstablishConn(cfg.Redis)
	pgConn := infra_pg_init.MustEstablishConn(cfg.Postgres)

	roomIDSet := infra_redis_roomid_set.New(redisConn, roomIDSetKey)

	roomRepo := infra_postgres_room.New(pgConn)
	voteRepo := infra_postgres_vote.New(pgConn)
	movieRepo := infra_postgres_meta.New(pgConn)

	embedder := infra_embedder_mock.New()
	embedderRepo := infra_qdrant_mock.New()

	roomUC := usecase_room.New(roomRepo, embedder, embedderRepo, roomIDSet)
	voteUC := usecase_vote.New(voteRepo)
	movieUC := usecase_movie.New(embedder, movieRepo, embedderRepo)

	hub := ws_room.New(slog.Default())

	controllerPool := http_init.NewControllerPool()

	controllerPool.Add(http_room.New(roomUC, hub))
	controllerPool.Add(http_voting.New(voteUC, hub))
	controllerPool.Add(http_movie.New(movieUC, hub))

	controllerPool.Register()
	controllerPool.RunAll(cfg.HTTP.Port)
}
