//go:build !integration
// +build !integration

package usecase_movie

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/humanbelnik/kinoswap/core/internal/service/embedding_reducer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	embedder_mocks "github.com/humanbelnik/kinoswap/core/internal/usecase/movie/mocks/movie/embedder"
	repo_mocks "github.com/humanbelnik/kinoswap/core/internal/usecase/movie/mocks/movie/repository"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
)

type UsecaseMovieUnitSuite struct {
	suite.Suite
}

type resources struct {
	usecase          *Usecase
	metaRepository   *repo_mocks.MetaRepository
	posterRepository *repo_mocks.PosterRepository
	embedder         *embedder_mocks.Embedder
	ctx              context.Context
	wg               sync.WaitGroup
}

type MovieMetaBuilder struct {
	mm model.MovieMeta
}

func NewMovieMetaBuilder() *MovieMetaBuilder {
	return &MovieMetaBuilder{
		mm: model.MovieMeta{
			ID:         uuid.New(),
			PosterLink: "http://example.com/poster.jpg",
			Title:      "Test Movie",
			Genres:     []string{"Drama", "Comedy"},
			Year:       2024,
			Rating:     8.5,
			Overview:   "Test overview",
		},
	}
}

func (b *MovieMetaBuilder) WithTitle(title string) *MovieMetaBuilder {
	b.mm.Title = title
	return b
}

func (b *MovieMetaBuilder) WithEmptyTitle() *MovieMetaBuilder {
	b.mm.Title = ""
	return b
}

func (b *MovieMetaBuilder) Build() model.MovieMeta {
	return b.mm
}

func initResources(t provider.T) *resources {
	metaRepository := repo_mocks.NewMetaRepository(t)
	posterRepository := repo_mocks.NewPosterRepository(t)
	embedder := embedder_mocks.NewEmbedder(t)
	embeddingReducer := embedding_reducer.New()
	usecase := New(metaRepository, posterRepository, embedder, embeddingReducer)

	return &resources{
		usecase:          usecase,
		metaRepository:   metaRepository,
		posterRepository: posterRepository,
		embedder:         embedder,
		ctx:              context.Background(),
	}
}

func (suite *UsecaseMovieUnitSuite) TestUpload(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, movie model.Movie)
		movie         model.Movie
		expectError   bool
		errorContains string
	}{
		{
			name: "Should upload movie successfully",
			setupMocks: func(r *resources, movie model.Movie) {
				r.metaRepository.On("Store", r.ctx, *movie.MM).Return(nil).Once()
				r.wg.Add(2)

				r.embedder.On("BuildMovieEmbedding", mock.Anything, *movie.MM).Return(model.Embedding(make([]float32, model.EmbeddingDimension)), nil).Once().Run(func(args mock.Arguments) {
					r.wg.Done()
				})
				r.metaRepository.On("StoreEmbedding", mock.Anything, movie.MM.ID, mock.AnythingOfType("model.Embedding")).Return(nil).Once().Run(func(args mock.Arguments) {
					r.wg.Done()
				})
			},
			movie: model.Movie{
				MM: func() *model.MovieMeta {
					mm := NewMovieMetaBuilder().Build()
					return &mm
				}(),
			},
			expectError: false,
		},
		{
			name: "Should return error when meta repository fails",
			setupMocks: func(r *resources, movie model.Movie) {
				repoError := errors.New("store error")
				r.metaRepository.On("Store", r.ctx, *movie.MM).Return(repoError).Once()
				r.metaRepository.On("Delete", r.ctx, movie.MM.ID).Return(nil).Once()
			},
			movie: model.Movie{
				MM: func() *model.MovieMeta {
					mm := NewMovieMetaBuilder().Build()
					return &mm
				}(),
			},
			expectError:   true,
			errorContains: "store error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r, tc.movie)

			err := r.usecase.Upload(r.ctx, tc.movie)
			r.wg.Wait()

			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
				assert.True(t, errors.Is(err, ErrInternal))
			} else {
				assert.NoError(t, err)
				r.metaRepository.AssertExpectations(t)
				r.embedder.AssertExpectations(t)
			}
		})
	}
}

func (suite *UsecaseMovieUnitSuite) TestDelete(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, movieID uuid.UUID)
		movieID       uuid.UUID
		expectError   bool
		errorContains string
	}{
		{
			name: "Should delete movie successfully",
			setupMocks: func(r *resources, movieID uuid.UUID) {
				r.metaRepository.On("Exists", r.ctx, movieID).Return(true, nil).Once()
				r.metaRepository.On("Delete", r.ctx, movieID).Return(nil).Once()
				r.posterRepository.On("Delete", r.ctx, movieID.String()).Return(nil).Once()
			},
			movieID:     uuid.New(),
			expectError: false,
		},
		{
			name: "Should return ErrResourceNotFound when movie doesn't exist",
			setupMocks: func(r *resources, movieID uuid.UUID) {
				r.metaRepository.On("Exists", r.ctx, movieID).Return(false, nil).Once()
			},
			movieID:     uuid.New(),
			expectError: true,
		},
		{
			name: "Should return error when meta repository delete fails",
			setupMocks: func(r *resources, movieID uuid.UUID) {
				r.metaRepository.On("Exists", r.ctx, movieID).Return(true, nil).Once()
				repoError := errors.New("delete error")
				r.metaRepository.On("Delete", r.ctx, movieID).Return(repoError).Once()
			},
			movieID:       uuid.New(),
			expectError:   true,
			errorContains: "delete error",
		},
		{
			name: "Should return error when exists check fails",
			setupMocks: func(r *resources, movieID uuid.UUID) {
				repoError := errors.New("exists error")
				r.metaRepository.On("Exists", r.ctx, movieID).Return(false, repoError).Once()
			},
			movieID:       uuid.New(),
			expectError:   true,
			errorContains: "exists error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r, tc.movieID)

			err := r.usecase.Delete(r.ctx, tc.movieID)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
					assert.True(t, errors.Is(err, ErrInternal))
				} else {
					assert.ErrorIs(t, err, ErrResourceNotFound)
				}
			} else {
				assert.NoError(t, err)
			}
			r.metaRepository.AssertExpectations(t)
			r.posterRepository.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseMovieUnitSuite) TestLoadAll(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupMocks     func(r *resources, expectedMovies []*model.MovieMeta)
		expectError    bool
		errorContains  string
		expectedMovies []*model.MovieMeta
	}{
		{
			name: "Should load movies successfully",
			setupMocks: func(r *resources, expectedMovies []*model.MovieMeta) {
				r.metaRepository.On("LoadAll", r.ctx).Return(expectedMovies, nil).Once()
				for _, movie := range expectedMovies {
					if movie.PosterLink != "" {
						r.posterRepository.On("GeneratePresignedURL", r.ctx, movie.PosterLink, mock.AnythingOfType("time.Duration")).Return("http://presigned.url/poster", nil).Once()
					}
				}
			},
			expectError: false,
			expectedMovies: func() []*model.MovieMeta {
				mm1 := NewMovieMetaBuilder().Build()
				mm2 := NewMovieMetaBuilder().Build()
				return []*model.MovieMeta{&mm1, &mm2}
			}(),
		},
		{
			name: "Should return ErrResourceNotFound when no movies exist",
			setupMocks: func(r *resources, expectedMovies []*model.MovieMeta) {
				r.metaRepository.On("LoadAll", r.ctx).Return([]*model.MovieMeta{}, nil).Once()
			},
			expectError:    true,
			expectedMovies: nil,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, expectedMovies []*model.MovieMeta) {
				repoError := errors.New("load error")
				r.metaRepository.On("LoadAll", r.ctx).Return(nil, repoError).Once()
			},
			expectError:    true,
			errorContains:  "load error",
			expectedMovies: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r, tc.expectedMovies)

			movies, err := r.usecase.LoadAll(r.ctx)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
					assert.True(t, errors.Is(err, ErrInternal))
				} else {
					assert.ErrorIs(t, err, ErrResourceNotFound)
				}
				assert.Nil(t, movies)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMovies, movies)
			}
			r.metaRepository.AssertExpectations(t)
			r.posterRepository.AssertExpectations(t)
		})
	}
}

func TestUnitSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseMovieUnitSuite))
}
