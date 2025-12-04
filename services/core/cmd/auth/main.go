package main

import (
	"github.com/humanbelnik/kinoswap/core/internal/config"
	http_auth "github.com/humanbelnik/kinoswap/core/internal/delivery/http/auth"
	http_init "github.com/humanbelnik/kinoswap/core/internal/delivery/http/init"
	infra_redis_init "github.com/humanbelnik/kinoswap/core/internal/infra/redis/init"
	infra_session_cache "github.com/humanbelnik/kinoswap/core/internal/infra/redis/session"
	servie_simple_auth "github.com/humanbelnik/kinoswap/core/internal/service/auth/simple"
)

func main() {
	cfg := config.Load()
	redisConn := infra_redis_init.MustEstablishConn(cfg.Redis)
	sessionCache := infra_session_cache.New(redisConn, "session_cache")
	authService := servie_simple_auth.New(nil, sessionCache, nil)
	controllerPool := http_init.NewControllerPool()
	controllerPool.Add(http_auth.New(authService))
	controllerPool.Register()
	controllerPool.RunAll("8080")
}
