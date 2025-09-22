package infra_postgres_meta

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Driver struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Driver {
	return &Driver{db: db}
}

func (d *Driver) Store(ctx context.Context, mm model.MovieMeta) error {
	const (
		q = `
		INSERT INTO movies (id, title, year, rating, genres, overview, poster_link)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		`
	)

	genresArr := pq.StringArray(mm.Genres)
	_, err := d.db.ExecContext(ctx, q,
		mm.ID,
		mm.Title,
		mm.Year,
		mm.Rating,
		genresArr,
		mm.Overview,
		mm.PosterLink,
	)

	if err != nil {
		return fmt.Errorf("failed to store movie meta: %w", err)
	}

	return nil
}

func (d *Driver) Load(ctx context.Context) ([]*model.MovieMeta, error) {
	const (
		q = `
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies
		ORDER BY title
		`
	)

	type movieRow struct {
		ID         uuid.UUID      `db:"id"`
		Title      string         `db:"title"`
		Year       int            `db:"year"`
		Rating     float64        `db:"rating"`
		Genres     pq.StringArray `db:"genres"`
		Overview   string         `db:"overview"`
		PosterLink string         `db:"poster_link"`
	}

	var rows []movieRow
	err := d.db.SelectContext(ctx, &rows, q)
	if err != nil {
		return nil, fmt.Errorf("failed to load movies: %w", err)
	}

	movies := make([]*model.MovieMeta, 0, len(rows))
	for _, row := range rows {
		movies = append(movies, &model.MovieMeta{
			ID:         row.ID,
			Title:      row.Title,
			Year:       row.Year,
			Rating:     row.Rating,
			Genres:     []string(row.Genres),
			Overview:   row.Overview,
			PosterLink: row.PosterLink,
		})
	}

	return movies, nil
}

func (d *Driver) LoadByID(ctx context.Context, ID uuid.UUID) (model.MovieMeta, error) {
	const (
		q = `
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies
		WHERE id = $1
		`
	)

	var row struct {
		ID         uuid.UUID      `db:"id"`
		Title      string         `db:"title"`
		Year       int            `db:"year"`
		Rating     float64        `db:"rating"`
		Genres     pq.StringArray `db:"genres"`
		Overview   string         `db:"overview"`
		PosterLink string         `db:"poster_link"`
	}

	err := d.db.GetContext(ctx, &row, q, ID)
	if err != nil {
		return model.MovieMeta{}, fmt.Errorf("failed to load movie by ID %s: %w", ID, err)
	}

	return model.MovieMeta{
		ID:         row.ID,
		Title:      row.Title,
		Year:       row.Year,
		Rating:     row.Rating,
		Genres:     []string(row.Genres),
		Overview:   row.Overview,
		PosterLink: row.PosterLink,
	}, nil
}

func (d *Driver) LoadByIDs(ctx context.Context, IDs []uuid.UUID) ([]*model.MovieMeta, error) {
	if len(IDs) == 0 {
		return []*model.MovieMeta{}, nil
	}

	placeholders := make([]string, len(IDs))
	args := make([]interface{}, len(IDs))
	for i, id := range IDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	q := fmt.Sprintf(`
		SELECT id, title, year, rating, genres, overview, poster_link
		FROM movies
		WHERE id IN (%s)
		ORDER BY title
	`, strings.Join(placeholders, ", "))

	type movieRow struct {
		ID         uuid.UUID      `db:"id"`
		Title      string         `db:"title"`
		Year       int            `db:"year"`
		Rating     float64        `db:"rating"`
		Genres     pq.StringArray `db:"genres"`
		Overview   string         `db:"overview"`
		PosterLink string         `db:"poster_link"`
	}

	var rows []movieRow
	err := d.db.SelectContext(ctx, &rows, q, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to load movies by IDs: %w", err)
	}

	movies := make([]*model.MovieMeta, 0, len(rows))
	for _, row := range rows {
		movies = append(movies, &model.MovieMeta{
			ID:         row.ID,
			Title:      row.Title,
			Year:       row.Year,
			Rating:     row.Rating,
			Genres:     []string(row.Genres),
			Overview:   row.Overview,
			PosterLink: row.PosterLink,
		})
	}

	return movies, nil
}

func (d *Driver) Update(ctx context.Context, mm model.MovieMeta) error {
	const (
		q = `
		UPDATE movies 
		SET 
			title = $2,
			year = $3,
			rating = $4,
			genres = $5,
			overview = $6,
			poster_link = $7
		WHERE id = $1
		`
	)

	genresArr := pq.StringArray(mm.Genres)
	result, err := d.db.ExecContext(ctx, q,
		mm.ID,
		mm.Title,
		mm.Year,
		mm.Rating,
		genresArr,
		mm.Overview,
		mm.PosterLink,
	)

	if err != nil {
		return fmt.Errorf("failed to update movie meta: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("movie with ID %s not found", mm.ID)
	}

	return nil
}

func (d *Driver) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	const (
		q = `DELETE FROM movies WHERE id = $1`
	)

	result, err := d.db.ExecContext(ctx, q, ID)
	if err != nil {
		return fmt.Errorf("failed to delete movie: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("movie with ID %s not found", ID)
	}

	return nil
}
