package infra_postgres_movie

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/jmoiron/sqlx"
)

var (
	ErrMovieNotFound    = errors.New("movie not found")
	ErrDuplicateMovie   = errors.New("movie already exists")
	ErrInvalidEmbedding = errors.New("invalid embedding dimensions")
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Store(ctx context.Context, mm model.MovieMeta) error {
	movieDB := FromDomain(mm)

	query := `
		INSERT INTO movies (id, title, year, rating, genres, overview, poster_link)
		VALUES (:id, :title, :year, :rating, :genres, :overview, :poster_link)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			year = EXCLUDED.year,
			rating = EXCLUDED.rating,
			genres = EXCLUDED.genres,
			overview = EXCLUDED.overview,
			poster_link = EXCLUDED.poster_link
	`

	_, err := r.db.NamedExecContext(ctx, query, movieDB)
	if err != nil {
		return fmt.Errorf("failed to store movie: %w", err)
	}

	return nil
}

func (r *Repository) Load(ctx context.Context) ([]*model.MovieMeta, error) {
	query := `
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies
	`

	var moviesDB []MovieDB
	err := r.db.SelectContext(ctx, &moviesDB, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query movies: %w", err)
	}

	movies := make([]*model.MovieMeta, len(moviesDB))
	for i, movieDB := range moviesDB {
		domainMovie := movieDB.ToDomain()
		movies[i] = &domainMovie
	}

	return movies, nil
}

func (r *Repository) LoadByID(ctx context.Context, ID uuid.UUID) (model.MovieMeta, error) {
	query := `
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies 
		WHERE id = $1
	`

	var movieDB MovieDB
	err := r.db.GetContext(ctx, &movieDB, query, ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.MovieMeta{}, ErrMovieNotFound
		}
		return model.MovieMeta{}, fmt.Errorf("failed to load movie by id: %w", err)
	}

	return movieDB.ToDomain(), nil
}

func (r *Repository) LoadByIDs(ctx context.Context, IDs []uuid.UUID) ([]*model.MovieMeta, error) {
	if len(IDs) == 0 {
		return []*model.MovieMeta{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies 
		WHERE id IN (?)
	`, IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	query = r.db.Rebind(query)
	var moviesDB []MovieDB
	err = r.db.SelectContext(ctx, &moviesDB, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query movies by ids: %w", err)
	}

	movies := make([]*model.MovieMeta, len(moviesDB))
	for i, movieDB := range moviesDB {
		domainMovie := movieDB.ToDomain()
		movies[i] = &domainMovie
	}

	return movies, nil
}

func (r *Repository) Update(ctx context.Context, mm model.MovieMeta) error {
	movieDB := FromDomain(mm)
	query := `
		UPDATE movies 
		SET title = :title, year = :year, rating = :rating, genres = :genres, 
			overview = :overview, poster_link = :poster_link
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, movieDB)
	if err != nil {
		return fmt.Errorf("failed to update movie: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMovieNotFound
	}

	return nil
}

func (r *Repository) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	query := `DELETE FROM movies WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, ID)
	if err != nil {
		return fmt.Errorf("failed to delete movie: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMovieNotFound
	}

	return nil
}

func (r *Repository) UpdateEmbedding(ctx context.Context, ID uuid.UUID, e model.Embedding) error {
	if len(e) != 384 {
		return ErrInvalidEmbedding
	}

	query := `UPDATE movies SET movie_vector = $2 WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, ID, e, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update embedding: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrMovieNotFound
	}

	return nil
}

func (r *Repository) KNN(ctx context.Context, k int, e model.Embedding) ([]*model.MovieMeta, error) {
	if len(e) != 384 {
		return nil, ErrInvalidEmbedding
	}

	if k <= 0 {
		return nil, errors.New("k must be positive")
	}

	query := `
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies 
		WHERE movie_vector IS NOT NULL
		ORDER BY movie_vector <=> $1
		LIMIT $2
	`

	var moviesDB []MovieDB
	err := r.db.SelectContext(ctx, &moviesDB, query, e, k)
	if err != nil {
		return nil, fmt.Errorf("failed to query KNN: %w", err)
	}

	movies := make([]*model.MovieMeta, len(moviesDB))
	for i, movieDB := range moviesDB {
		domainMovie := movieDB.ToDomain()
		movies[i] = &domainMovie
	}

	return movies, nil
}
