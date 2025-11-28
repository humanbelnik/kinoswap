package infra_postgres_vote

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	usecase_vote "github.com/humanbelnik/kinoswap/core/internal/usecase/vote"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

type Driver struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Driver {
	return &Driver{db: db}
}

type roomDTO struct {
	ID uuid.UUID `db:"id"`
}

type movieDTO struct {
	ID         uuid.UUID      `db:"id"`
	Title      string         `db:"title"`
	Year       int            `db:"year"`
	Rating     float64        `db:"rating"`
	Genres     pq.StringArray `db:"genres"` // Используем pq.StringArray для сканирования
	Overview   string         `db:"overview"`
	PosterLink string         `db:"poster_link"`
}

type resultDTO struct {
	MovieID    uuid.UUID      `db:"movie_id"`
	Title      string         `db:"title"`
	Year       int            `db:"year"`
	Rating     float64        `db:"rating"`
	Genres     pq.StringArray `db:"genres"` // Используем pq.StringArray для сканирования
	Overview   string         `db:"overview"`
	PosterLink string         `db:"poster_link"`
	Likes      int            `db:"likes"`
}

func (d *Driver) RoomIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	var room roomDTO

	query := `SELECT id FROM rooms WHERE code = $1`

	err := d.db.GetContext(ctx, &room, query, code)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, usecase_vote.ErrResourceNotFound
		}
		return uuid.Nil, err
	}

	return room.ID, nil
}

func (d *Driver) ParticipantsEmbeddings(ctx context.Context, roomID uuid.UUID) ([]model.Embedding, error) {
	var vectors []pgvector.Vector

	query := `
		SELECT preference 
		FROM participants 
		WHERE room_id = $1 AND preference IS NOT NULL
	`

	err := d.db.SelectContext(ctx, &vectors, query, roomID)
	if err != nil {
		return nil, err
	}

	embeddings := make([]model.Embedding, 0, len(vectors))
	for _, vector := range vectors {
		if len(vector.Slice()) > 0 {
			embeddings = append(embeddings, model.Embedding(vector.Slice()))
		}
	}

	return embeddings, nil
}

func (d *Driver) SimilarMovies(ctx context.Context, queryEmbedding []float32, limit int) ([]*model.MovieMeta, error) {
	var movies []movieDTO

	query := `
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies 
		WHERE movie_vector IS NOT NULL
		ORDER BY movie_vector <-> $1
		LIMIT $2
	`

	err := d.db.SelectContext(ctx, &movies, query, pgvector.NewVector(queryEmbedding), limit)
	if err != nil {
		return nil, err
	}

	result := make([]*model.MovieMeta, 0, len(movies))
	for _, movie := range movies {
		result = append(result, &model.MovieMeta{
			ID:         movie.ID,
			Title:      movie.Title,
			Year:       movie.Year,
			Rating:     movie.Rating,
			Genres:     []string(movie.Genres), // Конвертируем pq.StringArray в []string
			Overview:   movie.Overview,
			PosterLink: movie.PosterLink,
		})
	}

	return result, nil
}

func (d *Driver) Results(ctx context.Context, roomID uuid.UUID) ([]*model.Result, error) {
	var results []resultDTO

	query := `
		SELECT 
			m.id as movie_id, 
			m.title, 
			m.year, 
			m.rating, 
			m.genres, 
			m.overview, 
			m.poster_link,
			COALESCE(r.likes, 0) as likes
		FROM movies m
		LEFT JOIN reactions r ON m.id = r.movie_id AND r.room_id = $1
		WHERE r.likes > 0
		ORDER BY likes DESC
	`

	err := d.db.SelectContext(ctx, &results, query, roomID)
	if err != nil {
		return nil, err
	}

	modelResults := make([]*model.Result, 0, len(results))
	for _, r := range results {
		movieMeta := model.MovieMeta{
			ID:         r.MovieID,
			Title:      r.Title,
			Year:       r.Year,
			Rating:     r.Rating,
			Genres:     []string(r.Genres), // Конвертируем pq.StringArray в []string
			Overview:   r.Overview,
			PosterLink: r.PosterLink,
		}
		modelResults = append(modelResults, &model.Result{
			MM:    movieMeta,
			Likes: r.Likes,
		})
	}

	return modelResults, nil
}

func (d *Driver) AddReactions(ctx context.Context, roomID uuid.UUID, userID uuid.UUID, reactions map[uuid.UUID]int) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var voted bool
	checkParticipantQuery := `
		SELECT voted 
		FROM participants 
		WHERE id = $1 AND room_id = $2
	`

	err = tx.GetContext(ctx, &voted, checkParticipantQuery, userID, roomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return usecase_vote.ErrResourceNotFound
		}
		return err
	}

	if voted {
		return nil
	}

	if err := d.insertReactions(ctx, reactions, tx, roomID); err != nil {
		return err
	}

	updateVotedQuery := `
		UPDATE participants 
		SET voted = true 
		WHERE id = $1 AND room_id = $2
	`

	_, err = tx.ExecContext(ctx, updateVotedQuery, userID, roomID)
	if err != nil {
		return err
	}

	updateReadyQuery := `
		UPDATE rooms 
		SET ready = ready + 1
		WHERE id = $1
	`

	_, err = tx.ExecContext(ctx, updateReadyQuery, roomID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (d *Driver) insertReactions(ctx context.Context, reactions map[uuid.UUID]int, tx *sqlx.Tx, roomID uuid.UUID) error {
	for movieID, reaction := range reactions {
		if reaction == 1 {
			upsertQuery := `
				INSERT INTO reactions (id, room_id, movie_id, likes) 
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (room_id, movie_id) 
				DO UPDATE SET likes = reactions.likes + EXCLUDED.likes
			`

			_, err := tx.ExecContext(ctx, upsertQuery, uuid.New(), roomID, movieID, 1)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Driver) IsAllReady(ctx context.Context, roomID uuid.UUID) (bool, error) {
	var result struct {
		ReadyCount        int `db:"ready_count"`
		ParticipantsCount int `db:"participants_count"`
	}

	query := `
		SELECT 
			r.ready as ready_count,
			COUNT(p.id) as participants_count
		FROM rooms r
		LEFT JOIN participants p ON r.id = p.room_id
		WHERE r.id = $1
		GROUP BY r.id, r.ready
	`

	err := d.db.GetContext(ctx, &result, query, roomID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, usecase_vote.ErrResourceNotFound
		}
		return false, fmt.Errorf("failed to check readiness: %w", err)
	}

	return result.ReadyCount == result.ParticipantsCount, nil
}
