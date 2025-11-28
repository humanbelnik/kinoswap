package http_init

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	http_access_middleware "github.com/humanbelnik/kinoswap/core/internal/delivery/http/middleware/access"
)

const apiPrefix = "/api/v1"

type Controller interface {
	RegisterRoutes(router *gin.RouterGroup)
}

type ControllerPool struct {
	pool   []Controller
	rg     *gin.RouterGroup
	engine *gin.Engine
}

func NewControllerPool() *ControllerPool {
	engine := gin.Default() // ! Change on NGINX setup
	// engine.Use(cors.New(cors.Config{
	// 	AllowOrigins:     []string{"*"},
	// 	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	// 	AllowHeaders:     []string{"Origin", "Content-Type", "X-user-token", "Authorization"},
	// 	ExposeHeaders:    []string{"Content-Length", "X-user-token"},
	// 	AllowCredentials: false,
	// 	MaxAge:           12 * time.Hour,
	// }))

	mode := os.Getenv("MODE")
	engine.Use(http_access_middleware.ReadOnlyBadGatewayMiddleware(mode))
	rg := engine.Group(apiPrefix)

	return &ControllerPool{
		pool:   make([]Controller, 0, 10),
		rg:     rg,
		engine: engine,
	}
}

func (pool *ControllerPool) Register() {
	for _, c := range pool.pool {
		c.RegisterRoutes(pool.rg)
	}
}

func (pool *ControllerPool) RunAll(port string) {
	if err := pool.engine.Run(":" + port); err != nil {
		log.Fatalf("failed to run HTTP server: %v", err)
	}
}

func (pool *ControllerPool) Add(c Controller) {
	pool.pool = append(pool.pool, c)
}
