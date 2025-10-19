package infra_postgres_room

import (
	"context"
	"database/sql"
	"strings"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	usecase_room "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
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

type roomDTO struct {
	ID      uuid.UUID `db:"id"`
	IDAdmin uuid.UUID `db:"id_admin"`
	Code    string    `db:"code"`
	Status  string    `db:"status"`
}

func (d *Driver) CreateAndBook(ctx context.Context, room model.Room, ownerID uuid.UUID) error {
	roomDTO := roomDTO{
		ID:      room.ID,
		IDAdmin: ownerID,
		Code:    room.PublicCode,
		Status:  room.Status,
	}

	query := `
		INSERT INTO rooms (id, id_admin, code, status)
		VALUES (:id, :id_admin, :code, :status)
	`

	_, err := d.db.NamedExecContext(ctx, query, roomDTO)
	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") ||
			strings.Contains(err.Error(), "duplicate key") {
			return usecase_room.ErrCodeConflict
		}
		return err
	}
	return nil
}

func (d *Driver) IsOwner(ctx context.Context, code string, ownerID uuid.UUID) (bool, error) {
	var room roomDTO

	query := `
        SELECT id_admin 
        FROM rooms 
        WHERE code = $1
    `

	err := d.db.GetContext(ctx, &room, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, usecase_room.ErrResourceNotFound
		}
		return false, err
	}

	return room.IDAdmin == ownerID, nil
}

func (d *Driver) DeleteByCode(ctx context.Context, code string) error {
	query := `
        DELETE FROM rooms 
        WHERE code = $1
    `

	result, err := d.db.ExecContext(ctx, query, code)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return usecase_room.ErrResourceNotFound
	}

	return nil
}

func (d *Driver) StatusByCode(ctx context.Context, code string) (string, error) {
	var room roomDTO

	query := `
        SELECT status 
        FROM rooms 
        WHERE code = $1
    `

	err := d.db.GetContext(ctx, &room, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", usecase_room.ErrResourceNotFound
		}
		return "", err
	}

	return room.Status, nil
}

func (d *Driver) SetStatusByCode(ctx context.Context, code string, status string) error {
	query := `
        UPDATE rooms 
        SET status = $1 
        WHERE code = $2
    `

	result, err := d.db.ExecContext(ctx, query, status, code)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return usecase_room.ErrResourceNotFound
	}

	return nil
}

func (d *Driver) AddPreferenceEmbedding(ctx context.Context, code string, userID uuid.UUID, embedding model.Embedding) error {
	var roomID uuid.UUID
	queryGetRoomID := `SELECT id FROM rooms WHERE code = $1`

	err := d.db.GetContext(ctx, &roomID, queryGetRoomID, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return usecase_room.ErrResourceNotFound
		}
		return err
	}

	query := `
        INSERT INTO participants (id, room_id, preference) 
        VALUES ($1, $2, $3)
        ON CONFLICT (id) 
        DO UPDATE SET preference = $3
    `

	_, err = d.db.ExecContext(ctx, query, userID, roomID, pgvector.NewVector(embedding))

	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) ParticipantsCount(ctx context.Context, code string) (int, error) {
	var count int
	query := `
        SELECT COUNT(p.id) 
        FROM participants p
        JOIN rooms r ON p.room_id = r.id
        WHERE r.code = $1
    `

	err := d.db.GetContext(ctx, &count, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, usecase_room.ErrResourceNotFound
		}
		return 0, err
	}

	return count, nil
}

// type RoomStatus = string

// const (
// 	RoomStatusAcquired RoomStatus = "acquired"
// 	RoomStatusFree     RoomStatus = "free"
// )

// type Driver struct {
// 	db *sqlx.DB
// }

// func New(
// 	db *sqlx.DB,
// ) *Driver {
// 	return &Driver{db: db}
// }

// func (d *Driver) CreateAndAquire(ctx context.Context, roomID model.RoomID) error {
// 	const (
// 		q = `INSERT INTO rooms (id, status)
// 		VALUES ($1, $2)`
// 	)
// 	_, err := d.db.ExecContext(ctx, q,
// 		roomID,
// 		RoomStatusAcquired,
// 	)

// 	return err
// }

// func (d *Driver) FindAndAcquire(ctx context.Context) (model.RoomID, error) {
// 	const (
// 		q = `
// 		UPDATE rooms
// 		SET status = $1
// 		WHERE id = (
// 			SELECT id FROM rooms
// 			WHERE status = $2
// 			FOR UPDATE SKIP LOCKED
// 			LIMIT 1
// 		)
// 		RETURNING id`
// 	)
// 	var roomID string
// 	if err := d.db.QueryRowContext(ctx, q, RoomStatusAcquired, RoomStatusFree).Scan(&roomID); err != nil {
// 		return model.EmptyRoomID, err
// 	}

// 	return model.RoomID(roomID), nil
// }

// func (d *Driver) TryAcquire(ctx context.Context, roomID model.RoomID) error {
// 	const (
// 		q = `UPDATE rooms SET status = $1 WHERE id = $2 AND status = $3`
// 	)
// 	res, err := d.db.ExecContext(ctx, q, RoomStatusAcquired, string(roomID), RoomStatusFree)
// 	if err != nil {
// 		return err
// 	}
// 	rowsAffected, _ := res.RowsAffected()
// 	if rowsAffected == 0 {
// 		return fmt.Errorf("failed to acquire %s", roomID)
// 	}

// 	return nil
// }

// func (d *Driver) IsExistsRoomID(ctx context.Context, roomID model.RoomID) (bool, error) {
// 	query := `SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)`

// 	var exists bool
// 	err := d.db.QueryRowContext(ctx, query, string(roomID)).Scan(&exists)
// 	if err != nil {
// 		return false, err
// 	}

// 	return exists, nil
// }

// func (d *Driver) AppendPreference(ctx context.Context, roomID model.RoomID, preference model.Preference) error {
// 	const (
// 		q = `
// 		UPDATE rooms
// 		SET
// 			preferences = array_append(preferences, $1),
// 			participants = participants + 1
// 		WHERE id = $2 AND status = $3
// 		`
// 	)

// 	_, err := d.db.ExecContext(ctx, q, preference.Text, string(roomID), RoomStatusAcquired)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (d *Driver) IsRoomAcquired(ctx context.Context, roomID model.RoomID) (bool, error) {
// 	const (
// 		q = `SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1 AND status = $2)`
// 	)

// 	var exists bool
// 	err := d.db.QueryRowContext(ctx, q, string(roomID), RoomStatusAcquired).Scan(&exists)
// 	if err != nil {
// 		return false, err
// 	}

// 	return exists, nil
// }

// // Clear fields
// // Set status to Free
// // Add ID to a set of freeds
// func (d *Driver) ReleaseRoom(ctx context.Context, roomID model.RoomID) error {
// 	return nil
// }
