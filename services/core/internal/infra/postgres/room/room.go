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

func (d *Driver) IsParticipant(ctx context.Context, code string, userID uuid.UUID) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1 
			FROM participants p
			JOIN rooms r ON p.room_id = r.id
			WHERE r.code = $1 AND p.id = $2
		)
	`

	err := d.db.GetContext(ctx, &exists, query, code, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, usecase_room.ErrResourceNotFound
		}
		return false, err
	}

	return exists, nil
}

func (d *Driver) UUIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	var _uuid uuid.UUID

	query := `
		SELECT id FROM rooms WHERE code = $1
	`
	if err := d.db.GetContext(ctx, &_uuid, query, code); err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, usecase_room.ErrResourceNotFound
		}
		return uuid.Nil, err
	}
	return _uuid, nil
}
