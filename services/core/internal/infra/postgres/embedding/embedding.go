package infra_postgres_embedding

import (
	"context"

	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/jmoiron/sqlx"
)

type Driver struct {
	db *sqlx.DB
}

func New(
	db *sqlx.DB,
) *Driver {
	return &Driver{db: db}
}

func (d *Driver) Append(ctx context.Context, roomID model.RoomID, E model.Embedding) error {
	_, err := d.db.ExecContext(ctx,
		`UPDATE rooms 
		 SET preferences_vec = array_append(preferences_vec, $1)
		 WHERE id = $2`,
		E,
		string(roomID),
	)
	return err
}
