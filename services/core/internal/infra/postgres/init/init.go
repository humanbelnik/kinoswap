package infra_pg_init

import (
	"fmt"
	"log"

	"github.com/humanbelnik/kinoswap/core/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func MustEstablishConn(cfg config.Postgres) *sqlx.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode,
	)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}
