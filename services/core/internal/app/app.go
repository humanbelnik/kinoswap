package app

import (
	"github.com/humanbelnik/kinoswap/core/internal/config"
	http_auth "github.com/humanbelnik/kinoswap/core/internal/delivery/http/auth"
	http_init "github.com/humanbelnik/kinoswap/core/internal/delivery/http/init"
	http_auth_middleware "github.com/humanbelnik/kinoswap/core/internal/delivery/http/middleware/auth"
	http_movie "github.com/humanbelnik/kinoswap/core/internal/delivery/http/movie"
	http_room "github.com/humanbelnik/kinoswap/core/internal/delivery/http/room"
	http_swagger "github.com/humanbelnik/kinoswap/core/internal/delivery/http/swagger"
	ws_room "github.com/humanbelnik/kinoswap/core/internal/delivery/ws/room"
	infra_embedder "github.com/humanbelnik/kinoswap/core/internal/infra/embedder"
	infra_pg_init "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/init"
	infra_postgres_movie "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/movie"
	infra_postgres_room "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/room"
	infra_redis_init "github.com/humanbelnik/kinoswap/core/internal/infra/redis/init"
	infra_session_cache "github.com/humanbelnik/kinoswap/core/internal/infra/redis/session"
	infra_s3 "github.com/humanbelnik/kinoswap/core/internal/infra/s3"
	servie_simple_auth "github.com/humanbelnik/kinoswap/core/internal/service/auth/simple"
	"github.com/humanbelnik/kinoswap/core/internal/service/embedding_reducer"
	usecase_movie "github.com/humanbelnik/kinoswap/core/internal/usecase/movie"
	usecase_room "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
)

func Go(cfg *config.Config) {

	redisConn := infra_redis_init.MustEstablishConn(cfg.Redis)
	pgConn := infra_pg_init.MustEstablishConn(cfg.Postgres)
	embedder := infra_embedder.MustEstablishConnection(cfg.Embedder)
	s3conn := infra_s3.MustEstabilishConn()

	posterRepository, err := infra_s3.New("hbk-test-bucket", s3conn, "poster/")
	if err != nil {
		panic(err)
	}

	embeddingReducer := embedding_reducer.New()

	// roomIDSet := infra_redis_roomid_set.New(redisConn, roomIDSetKey)

	roomRepository := infra_postgres_room.New(pgConn)
	// voteRepo := infra_postgres_vote.New(pgConn)
	movieRepository := infra_postgres_movie.New(pgConn)
	// embedderRepo := infra_postgres_embedding.New(pgConn)

	roomUC := usecase_room.New(roomRepository, embedder)

	hub := ws_room.NewHub(roomUC)
	go hub.Run()
	// voteUC := usecase_vote.New(voteRepo)
	movieUC := usecase_movie.New(movieRepository, posterRepository, embedder, embeddingReducer)

	sessionCache := infra_session_cache.New(redisConn, "session_cache")
	authService := servie_simple_auth.New(nil, sessionCache, nil)
	authMiddleware := http_auth_middleware.New(authService)

	controllerPool := http_init.NewControllerPool()
	controllerPool.Add(http_swagger.New())
	controllerPool.Add(http_room.New(roomUC))
	controllerPool.Add(http_movie.New(movieUC, authMiddleware))
	controllerPool.Add(http_auth.New(authService))
	controllerPool.Add(ws_room.NewController(hub))

	controllerPool.Register()
	controllerPool.RunAll(cfg.HTTP.Port)
}
