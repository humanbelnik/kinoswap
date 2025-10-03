package http_init

import (
	"log"

	"github.com/gin-gonic/gin"
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
