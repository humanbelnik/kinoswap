//go:build integration
// +build integration

package integrationtest

import (
	"context"
	"testing"

	"github.com/google/uuid"
	infra_pg_init "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/init"
	infra_postgres_movie "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/movie"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/humanbelnik/kinoswap/core/internal/service/embedding_reducer"
	movie_usecase "github.com/humanbelnik/kinoswap/core/internal/usecase/movie"
	mocks_embedder "github.com/humanbelnik/kinoswap/core/internal/usecase/movie/mocks/movie/embedder"
	mocks_s3 "github.com/humanbelnik/kinoswap/core/internal/usecase/movie/mocks/movie/repository"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMovieIntegrationSuite struct {
	suite.Suite
	uc *movie_usecase.Usecase
}

func initUsecase(t provider.T) *movie_usecase.Usecase {
	cfg := getConfig()

	pgConn := infra_pg_init.MustEstablishConn(cfg.Postgres)
	posterRepository := mocks_s3.NewPosterRepository(t)

	embeddingReducer := embedding_reducer.New()
	movieRepository := infra_postgres_movie.New(pgConn)
	embedder := mocks_embedder.NewEmbedder(t)

	return movie_usecase.New(movieRepository, posterRepository, embedder, embeddingReducer)
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
				s.uc.Embedder = embedderConn
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
				s.uc.MetaRepository.Delete(ctx, movieID)
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
				s.uc.Embedder = embedderConn
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
				s.uc.MetaRepository.Delete(ctx, movieID)
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
