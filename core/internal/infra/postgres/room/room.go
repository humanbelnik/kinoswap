package infra_postgres_room

import (
	"context"
	"fmt"

	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/jmoiron/sqlx"
)

type RoomStatus = string

const (
	RoomStatusAcquired RoomStatus = "acquired"
	RoomStatusFree     RoomStatus = "free"
)

type Driver struct {
	db *sqlx.DB
}

func New(
	db *sqlx.DB,
) *Driver {
	return &Driver{db: db}
}

func (d *Driver) CreateAndAquire(ctx context.Context, roomID model.RoomID) error {
	const (
		q = `INSERT INTO rooms (id, status)
		VALUES ($1, $2)`
	)
	_, err := d.db.ExecContext(ctx, q,
		roomID,
		RoomStatusAcquired,
	)

	return err
}

func (d *Driver) FindAndAcquire(ctx context.Context) (model.RoomID, error) {
	const (
		q = `
		UPDATE rooms 
		SET status = $1
		WHERE id = (
			SELECT id FROM rooms 
			WHERE status = $2
			ORDER BY created_at ASC 
			FOR UPDATE SKIP LOCKED 
			LIMIT 1
		)
		RETURNING id`
	)
	var roomID string
	if err := d.db.QueryRowContext(ctx, q, RoomStatusAcquired, RoomStatusFree).Scan(&roomID); err != nil {
		return model.EmptyRoomID, err
	}

	return model.RoomID(roomID), nil
}

func (d *Driver) TryAcquire(ctx context.Context, roomID model.RoomID) error {
	const (
		q = `UPDATE rooms SET status = $1 WHERE id = $2 AND status = $3`
	)
	res, err := d.db.ExecContext(ctx, q, RoomStatusAcquired, string(roomID), RoomStatusFree)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("failed to acquire %s", roomID)
	}

	return nil
}

func (d *Driver) IsExistsRoomID(ctx context.Context, roomID model.RoomID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)`

	var exists bool
	err := d.db.QueryRowContext(ctx, query, string(roomID)).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
