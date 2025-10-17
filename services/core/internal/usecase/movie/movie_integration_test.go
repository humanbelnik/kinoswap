//go:build integration

package usecase_movie

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/config"
	infra_postgres_embedding "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/embedding"
	infra_pg_init "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/init"
	infra_postgres_movie "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/movie"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/humanbelnik/kinoswap/core/internal/service/embedding_reducer"
	mocks_embedder "github.com/humanbelnik/kinoswap/core/mocks/movie/embedder"
	mocks_s3 "github.com/humanbelnik/kinoswap/core/mocks/movie/s3"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMovieIntegrationSuite struct {
	suite.Suite
	uc *Usecase[*model.Poster]
}

func initUsecase(t provider.T) *Usecase[*model.Poster] {
	cfg := config.Load()

	pgConn := infra_pg_init.MustEstablishConn(cfg.Postgres)
	posterStorage := mocks_s3.NewPosterStorage[*model.Poster](t)

	embeddingReducer := embedding_reducer.New()
	movieRepo := infra_postgres_movie.New(pgConn)
	embedderRepo := infra_postgres_embedding.New(pgConn)
	embedderConn := mocks_embedder.NewEmbedder(t)

	return New(posterStorage, embedderConn, movieRepo, embedderRepo, embeddingReducer)
}

func (s *UsecaseMovieIntegrationSuite) BeforeAll(t provider.T) {
	s.uc = initUsecase(t)
}

func (s *UsecaseMovieIntegrationSuite) TestIntegrationExists(t provider.T) {
	ctx := context.Background()
	tt := []struct {
		name     string
		movie    model.Movie
		setup    func(movieID uuid.UUID)
		teardown func(movieID uuid.UUID)

		expectError   bool
		errorType     error
		errorContains string
	}{
		{
			name: "load movies from non-empty database",
			setup: func(movieID uuid.UUID) {

				embedderConn := mocks_embedder.NewEmbedder(t)
				embedderConn.On("BuildMovieEmbedding", mock.Anything, mock.Anything).
					Return(model.Embedding(make([]float32, 384)), nil)
				s.uc.embedder = embedderConn
				movie := model.Movie{
					MM: &model.MovieMeta{
						ID:         movieID,
						PosterLink: "",
						Title:      "Test name",
						Genres:     []string{"Drama", "Romance"},
						Year:       2023,
						Rating:     8.2,
						Overview:   "A test movie for integration testing without poster",
					},
					Poster: nil,
				}
				s.uc.Upload(context.Background(), movie)
			},
			teardown: func(movieID uuid.UUID) {
				s.uc.metaRepository.DeleteByID(ctx, movieID)
			},

			expectError: false,
		},
	}
	for _, test := range tt {
		t.Run(test.name, func(t provider.T) {
			t.Parallel()

			movieID := uuid.New()
			test.setup(movieID)
			_, err := s.uc.Exists(ctx, movieID)

			if test.expectError {
				assert.ErrorIs(t, err, test.errorType)
			} else {
				assert.NoError(t, err)
			}
			test.teardown(movieID)
		})
	}
}

func (s *UsecaseMovieIntegrationSuite) TestIntegrationUpload(t provider.T) {
	ctx := context.Background()
	tt := []struct {
		name     string
		movie    model.Movie
		setup    func(movieID uuid.UUID) model.Movie
		teardown func(movieID uuid.UUID)

		expectError   bool
		errorType     error
		errorContains string
	}{
		{
			name: "successfully upload movie without poster",
			setup: func(movieID uuid.UUID) model.Movie {

				embedderConn := mocks_embedder.NewEmbedder(t)
				embedderConn.On("BuildMovieEmbedding", mock.Anything, mock.Anything).
					Return(model.Embedding(make([]float32, 384)), nil)
				s.uc.embedder = embedderConn
				return model.Movie{
					MM: &model.MovieMeta{
						ID:         movieID,
						PosterLink: "",
						Title:      "Test name",
						Genres:     []string{"Drama", "Romance"},
						Year:       2023,
						Rating:     8.2,
						Overview:   "A test movie for integration testing without poster",
					},
					Poster: nil,
				}
			},
			teardown: func(movieID uuid.UUID) {
				s.uc.metaRepository.DeleteByID(ctx, movieID)
			},

			expectError: false,
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t provider.T) {
			t.Parallel()

			movieID := uuid.New()
			movie := test.setup(movieID)
			err := s.uc.Upload(ctx, movie)

			if test.expectError {
				assert.ErrorIs(t, err, test.errorType)
			} else {
				assert.NoError(t, err)
			}
			test.teardown(movieID)
		})
	}
}

func TestIntegrationSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseMovieIntegrationSuite))
}
