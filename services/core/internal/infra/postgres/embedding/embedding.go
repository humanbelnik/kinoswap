package infra_postgres_embedding

import (
	"context"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/pgvector/pgvector-go"
)

type Driver struct {
	db *sqlx.DB
}

func New(
	db *sqlx.DB,
) *Driver {
	return &Driver{db: db}
}

func (d *Driver) Store(ctx context.Context, movieID uuid.UUID, e model.Embedding) error {
	v := pgvector.NewVector(e)
	_, err := d.db.ExecContext(ctx,
		`UPDATE movies
	SET movie_vector = $1
	WHERE id = $2`, v, movieID,
	)

	return err
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
