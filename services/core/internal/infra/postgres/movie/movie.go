package infra_postgres_movie

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/pgvector/pgvector-go"
)

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (d *Repository) StoreEmbedding(ctx context.Context, movieID uuid.UUID, e model.Embedding) error {
	v := pgvector.NewVector(e)
	_, err := d.db.ExecContext(ctx,
		`UPDATE movies
	SET movie_vector = $1
	WHERE id = $2`, v, movieID,
	)

	return err
}

func (r *Repository) Store(ctx context.Context, mm model.MovieMeta) error {
	movieDB := FromDomain(mm)

	query := `
		INSERT INTO movies (id, title, year, rating, genres, overview, poster_link)
		VALUES (:id, :title, :year, :rating, :genres, :overview, :poster_link)
	`

	_, err := r.db.NamedExecContext(ctx, query, movieDB)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) LoadAll(ctx context.Context) ([]*model.MovieMeta, error) {
	query := `
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies
	`

	var moviesDB []MovieDB
	err := r.db.SelectContext(ctx, &moviesDB, query)
	if err != nil {
		return nil, err
	}

	movies := make([]*model.MovieMeta, len(moviesDB))
	for i, movieDB := range moviesDB {
		domainMovie := movieDB.ToDomain()
		movies[i] = &domainMovie
	}

	return movies, nil
}

func (r *Repository) LoadSome(ctx context.Context, ids []uuid.UUID) ([]*model.MovieMeta, error) {
	query := `
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies WHERE id = ANY($1)
	`

	var moviesDB []MovieDB
	err := r.db.SelectContext(ctx, &moviesDB, query, ids)
	if err != nil {
		return nil, err
	}

	movies := make([]*model.MovieMeta, len(moviesDB))
	for i, movieDB := range moviesDB {
		domainMovie := movieDB.ToDomain()
		movies[i] = &domainMovie
	}

	return movies, nil
}

func (r *Repository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM movies WHERE id = $1", id,
	).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repository) Delete(ctx context.Context, ID uuid.UUID) error {
	query := `DELETE FROM movies WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) KNN(ctx context.Context, k int, e model.Embedding) ([]*model.MovieMeta, error) {
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
