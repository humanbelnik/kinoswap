package integrationtest

import (
	"sync"

	"github.com/humanbelnik/kinoswap/core/internal/config"
)

var (
	cfg     *config.Config
	cfgOnce sync.Once
)

func getConfig() *config.Config {
	cfgOnce.Do(func() {
		cfg = config.Load()
	})
	return cfg
}
