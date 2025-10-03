package infra_postgres_vote

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Driver struct {
	db *sqlx.DB
}

func New(
	db *sqlx.DB,
) *Driver {
	return &Driver{db: db}
}
func (d *Driver) AddVote(ctx context.Context, roomID model.RoomID, voteResult model.VoteResult) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for movieMeta, reaction := range voteResult.Results {
		if reaction == model.PassReaction {
			const (
				q = `
				INSERT INTO results (room_id, movie_id, pass_count)
				VALUES ($1, $2, 1)
				ON CONFLICT (room_id, movie_id) 
				DO UPDATE SET pass_count = results.pass_count + 1
				`
			)

			_, err := tx.ExecContext(ctx, q, string(roomID), movieMeta.ID)
			if err != nil {
				return fmt.Errorf("failed to increment pass_count for movie %s: %w", movieMeta.ID, err)
			}
		}
	}

	return tx.Commit()
}

func (d *Driver) LoadResults(ctx context.Context, roomID model.RoomID) ([]*model.MovieMeta, error) {
	const (
		q = `
		SELECT 
			m.id, 
			m.title, 
			m.year, 
			m.rating, 
			m.genres, 
			m.overview, 
			m.poster_link,
			r.pass_count
		FROM results r
		JOIN movies m ON r.movie_id = m.id
		WHERE r.room_id = $1
		ORDER BY r.pass_count DESC
		`
	)

	type resultRow struct {
		ID         string         `db:"id"`
		Title      string         `db:"title"`
		Year       int            `db:"year"`
		Rating     float64        `db:"rating"`
		Genres     pq.StringArray `db:"genres"`
		Overview   string         `db:"overview"`
		PosterLink string         `db:"poster_link"`
		PassCount  int            `db:"pass_count"`
	}

	var rows []resultRow
	err := d.db.SelectContext(ctx, &rows, q, string(roomID))
	if err != nil {
		return nil, fmt.Errorf("failed to load results: %w", err)
	}

	movies := make([]*model.MovieMeta, 0, len(rows))
	for _, row := range rows {
		movieID, err := uuid.Parse(row.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid movie UUID: %w", err)
		}

		movies = append(movies, &model.MovieMeta{
			ID:         movieID,
			Title:      row.Title,
			Year:       row.Year,
			Rating:     row.Rating,
			Genres:     row.Genres,
			Overview:   row.Overview,
			PosterLink: row.PosterLink,
		})
	}

	return movies, nil
}
