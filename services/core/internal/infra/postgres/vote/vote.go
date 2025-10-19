package infra_postgres_vote

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	usecase_vote "github.com/humanbelnik/kinoswap/core/internal/usecase/vote"
	"github.com/jmoiron/sqlx"
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
	ID         uuid.UUID `db:"id"`
	Title      string    `db:"title"`
	Year       int       `db:"year"`
	Rating     float64   `db:"rating"`
	Genres     []string  `db:"genres"`
	Overview   string    `db:"overview"`
	PosterLink string    `db:"poster_link"`
}

type resultDTO struct {
	MovieID    uuid.UUID `db:"movie_id"`
	Title      string    `db:"title"`
	Year       int       `db:"year"`
	Rating     float64   `db:"rating"`
	Genres     []string  `db:"genres"`
	Overview   string    `db:"overview"`
	PosterLink string    `db:"poster_link"`
	Likes      int       `db:"likes"`
}

func (d *Driver) GetRoomIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
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

func (d *Driver) GetParticipantsEmbeddings(ctx context.Context, roomID uuid.UUID) ([]model.Embedding, error) {
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
func (d *Driver) GetSimilarMovies(ctx context.Context, queryEmbedding []float32, limit int) ([]*model.MovieMeta, error) {
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
			Genres:     movie.Genres,
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
			Genres:     r.Genres,
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

func (d *Driver) AddReactions(ctx context.Context, roomID uuid.UUID, reactions map[uuid.UUID]int) error {
	tx, err := d.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for movieID, likes := range reactions {
		if likes <= 0 {
			continue
		}

		// Используем UPSERT для добавления или обновления реакций
		query := `
			INSERT INTO reactions (id, room_id, movie_id, likes) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (room_id, movie_id) 
			DO UPDATE SET likes = reactions.likes + EXCLUDED.likes
		`

		_, err = tx.ExecContext(ctx, query, uuid.New(), roomID, movieID, likes)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
