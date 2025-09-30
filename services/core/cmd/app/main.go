package main

import (
	"github.com/humanbelnik/kinoswap/core/internal/app"
	"github.com/humanbelnik/kinoswap/core/internal/config"
)

func main() {
	app.Go(config.Load())
}
