package main

import (
	_ "github.com/humanbelnik/kinoswap/core/docs"
	"github.com/humanbelnik/kinoswap/core/internal/app"
	"github.com/humanbelnik/kinoswap/core/internal/config"
)

// @title Kinoswap API
// @version 1.0
// @host localhost:8080
// @BasePath /api/v1/

// @securityDefinitions.apikey AdminToken
// @in header
// @name X-admin-token

func main() {
	app.Go(config.Load())
}
