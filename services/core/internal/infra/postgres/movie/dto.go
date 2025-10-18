package infra_postgres_movie

import (
	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/lib/pq"
)

type MovieDB struct {
	ID          uuid.UUID       `db:"id"`
	PosterLink  string          `db:"poster_link"`
	Title       string          `db:"title"`
	Genres      pq.StringArray  `db:"genres"`
	Year        int             `db:"year"`
	Rating      float64         `db:"rating"`
	Overview    string          `db:"overview"`
	MovieVector model.Embedding `db:"movie_vector"`
}

func (m *MovieDB) ToDomain() model.MovieMeta {
	return model.MovieMeta{
		ID:         m.ID,
		PosterLink: m.PosterLink,
		Title:      m.Title,
		Genres:     []string(m.Genres),
		Year:       m.Year,
		Rating:     m.Rating,
		Overview:   m.Overview,
	}
}

func FromDomain(mm model.MovieMeta) MovieDB {
	return MovieDB{
		ID:         mm.ID,
		PosterLink: mm.PosterLink,
		Title:      mm.Title,
		Genres:     pq.StringArray(mm.Genres),
		Year:       mm.Year,
		Rating:     mm.Rating,
		Overview:   mm.Overview,
	}
}
