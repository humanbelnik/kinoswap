package app

import (
	"log/slog"

	"github.com/humanbelnik/kinoswap/core/internal/config"
	http_init "github.com/humanbelnik/kinoswap/core/internal/delivery/http/init"
	http_movie "github.com/humanbelnik/kinoswap/core/internal/delivery/http/movie"
	http_swagger "github.com/humanbelnik/kinoswap/core/internal/delivery/http/swagger"
	ws_room "github.com/humanbelnik/kinoswap/core/internal/delivery/ws/room"
	infra_embedder "github.com/humanbelnik/kinoswap/core/internal/infra/embedder"
	infra_pg_init "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/init"
	infra_postgres_movie "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/movie"
	infra_s3 "github.com/humanbelnik/kinoswap/core/internal/infra/s3"
	"github.com/humanbelnik/kinoswap/core/internal/service/embedding_reducer"
	usecase_movie "github.com/humanbelnik/kinoswap/core/internal/usecase/movie"
)

func Go(cfg *config.Config) {
	const (
		roomIDSetKey = "room_id"
	)

	//redisConn := infra_redis_init.MustEstablishConn(cfg.Redis)
	pgConn := infra_pg_init.MustEstablishConn(cfg.Postgres)
	embedder := infra_embedder.MustEstablishConnection(cfg.Embedder)
	s3conn := infra_s3.MustEstabilishConn()

	posterRepository, err := infra_s3.New("hbk-test-bucket", s3conn, "poster/")
	if err != nil {
		panic(err)
	}

	embeddingReducer := embedding_reducer.New()

	// roomIDSet := infra_redis_roomid_set.New(redisConn, roomIDSetKey)

	// roomRepo := infra_postgres_room.New(pgConn)
	// voteRepo := infra_postgres_vote.New(pgConn)
	movieRepository := infra_postgres_movie.New(pgConn)
	// embedderRepo := infra_postgres_embedding.New(pgConn)

	// roomUC := usecase_room.New(roomRepo, embedderConn, embedderRepo, roomIDSet)
	// voteUC := usecase_vote.New(voteRepo)
	movieUC := usecase_movie.New(movieRepository, posterRepository, embedder, embeddingReducer)

	hub := ws_room.New(slog.Default())

	controllerPool := http_init.NewControllerPool()

	controllerPool.Add(http_swagger.New())
	// controllerPool.Add(http_room.New(roomUC, hub))
	// controllerPool.Add(http_voting.New(voteUC, hub))
	controllerPool.Add(http_movie.New(movieUC, hub))

	controllerPool.Register()
	controllerPool.RunAll(cfg.HTTP.Port)
}
