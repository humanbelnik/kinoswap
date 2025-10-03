package usecase_movie

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	"github.com/humanbelnik/kinoswap/core/internal/service/embedding_reducer"
	embedder_mocks "github.com/humanbelnik/kinoswap/core/mocks/movie/embedder"
	repo_mocks "github.com/humanbelnik/kinoswap/core/mocks/movie/repository"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseMovieUnitSuite struct {
	suite.Suite
}

type resources struct {
	usecase             *Usecase
	repository          *repo_mocks.MovieRepository
	embeddingRepository *repo_mocks.EmbeddingRepository
	embedder            *embedder_mocks.Embedder
	ctx                 context.Context
	wg                  sync.WaitGroup
}

func validRoomID() model.RoomID {
	return model.RoomID("123456")
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
	b.mm.Title = model.EmptyTitle
	return b
}

func (b *MovieMetaBuilder) Build() model.MovieMeta {
	return b.mm
}

func initResources(t provider.T) *resources {
	repository := repo_mocks.NewMovieRepository(t)
	embeddingRepository := repo_mocks.NewEmbeddingRepository(t)
	embedder := embedder_mocks.NewEmbedder(t)
	reducer := embedding_reducer.New()
	usecase := New(embedder, repository, embeddingRepository, reducer)

	return &resources{
		usecase:             usecase,
		repository:          repository,
		embeddingRepository: embeddingRepository,
		embedder:            embedder,
		ctx:                 context.Background(),
	}
}

func (suite *UsecaseMovieUnitSuite) TestUpload(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, movieMeta model.MovieMeta)
		movieMeta     model.MovieMeta
		expectError   bool
		errorType     error
		errorContains string
	}{
		{
			name: "Should upload movie successfully",
			setupMocks: func(r *resources, movieMeta model.MovieMeta) {
				r.repository.On("Store", r.ctx, movieMeta).Return(nil).Once()
				r.wg.Add(2)

				r.embedder.On("BuildMovieEmbedding", mock.Anything, movieMeta).Return(model.Embedding{}, nil).Once().Run(func(args mock.Arguments) {
					r.wg.Done()
				})
				r.embeddingRepository.On("Store", mock.Anything, movieMeta.ID, mock.AnythingOfType("model.Embedding")).Return(nil).Once().Run(func(args mock.Arguments) {
					r.wg.Done()
				})
			},
			movieMeta:   NewMovieMetaBuilder().Build(),
			expectError: false,
		},
		{
			name:        "Should return ErrInvalidInput when title is empty",
			setupMocks:  func(r *resources, movieMeta model.MovieMeta) {},
			movieMeta:   NewMovieMetaBuilder().WithEmptyTitle().Build(),
			expectError: true,
			errorType:   ErrInvalidInput,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r, tc.movieMeta)

			err := r.usecase.Upload(r.ctx, tc.movieMeta)
			r.wg.Wait()

			if tc.expectError {
				assert.ErrorIs(t, err, tc.errorType)
			} else {
				assert.NoError(t, err)
				r.repository.AssertExpectations(t)
				r.embedder.AssertExpectations(t)
				r.embeddingRepository.AssertExpectations(t)
			}
		})
	}
}

func (suite *UsecaseMovieUnitSuite) TestDeleteMovie(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, movieID uuid.UUID)
		movieID       uuid.UUID
		expectError   bool
		errorType     error
		errorContains string
	}{
		{
			name: "Should delete movie successfully",
			setupMocks: func(r *resources, movieID uuid.UUID) {
				r.repository.On("DeleteByID", r.ctx, movieID).Return(nil).Once()
			},
			movieID:     uuid.New(),
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, movieID uuid.UUID) {
				repoError := errors.New("delete error")
				r.repository.On("DeleteByID", r.ctx, movieID).Return(repoError).Once()
			},
			movieID:       uuid.New(),
			expectError:   true,
			errorType:     ErrFailedToDeleteMeta,
			errorContains: "delete error",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r, tc.movieID)

			err := r.usecase.DeleteMovie(r.ctx, tc.movieID)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.errorType)
				assert.ErrorContains(t, err, tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
			r.repository.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseMovieUnitSuite) TestUpdateMovie(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, movieMeta model.MovieMeta)
		movieMeta     model.MovieMeta
		expectError   bool
		errorType     error
		errorContains string
	}{
		{
			name: "Should update movie successfully",
			setupMocks: func(r *resources, movieMeta model.MovieMeta) {
				r.repository.On("Update", r.ctx, movieMeta).Return(nil).Once()
				r.wg.Add(2)
				r.embedder.On("BuildMovieEmbedding", mock.Anything, movieMeta).Return(model.Embedding{}, nil).Once().Run(func(args mock.Arguments) {
					r.wg.Done()
				})
				r.embeddingRepository.On("Store", mock.Anything, movieMeta.ID, mock.AnythingOfType("model.Embedding")).Return(nil).Once().Run(func(args mock.Arguments) {
					r.wg.Done()
				})
			},
			movieMeta:   NewMovieMetaBuilder().Build(),
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, movieMeta model.MovieMeta) {
				repoError := errors.New("update error")
				r.repository.On("Update", r.ctx, movieMeta).Return(repoError).Once()
			},
			movieMeta:     NewMovieMetaBuilder().Build(),
			expectError:   true,
			errorType:     ErrFailedToUpdateMeta,
			errorContains: "update error",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r, tc.movieMeta)

			err := r.usecase.UpdateMovie(r.ctx, tc.movieMeta)

			r.wg.Wait()
			if tc.expectError {
				assert.ErrorIs(t, err, tc.errorType)
				assert.ErrorContains(t, err, tc.errorContains)
			} else {
				assert.NoError(t, err)
				r.repository.AssertExpectations(t)
				r.embedder.AssertExpectations(t)
				r.embeddingRepository.AssertExpectations(t)
			}
		})
	}
}

func (suite *UsecaseMovieUnitSuite) TestLoad(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupMocks     func(r *resources, expectedMovies []*model.MovieMeta)
		expectError    bool
		errorType      error
		errorContains  string
		expectedMovies []*model.MovieMeta
	}{
		{
			name: "Should load movies successfully",
			setupMocks: func(r *resources, expectedMovies []*model.MovieMeta) {
				r.repository.On("Load", r.ctx).Return(expectedMovies, nil).Once()
			},
			expectError: false,
			expectedMovies: func() []*model.MovieMeta {
				expectedMoviesValue := []model.MovieMeta{
					NewMovieMetaBuilder().Build(),
					NewMovieMetaBuilder().Build(),
				}
				return []*model.MovieMeta{&expectedMoviesValue[0], &expectedMoviesValue[1]}
			}(),
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, expectedMovies []*model.MovieMeta) {
				repoError := errors.New("load error")
				r.repository.On("Load", r.ctx).Return(nil, repoError).Once()
			},
			expectError:    true,
			errorType:      ErrFailedToLoadMeta,
			errorContains:  "load error",
			expectedMovies: nil,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r, tc.expectedMovies)

			movies, err := r.usecase.Load(r.ctx)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.errorType)
				assert.ErrorContains(t, err, tc.errorContains)
				assert.Nil(t, movies)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMovies, movies)
			}
			r.repository.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseMovieUnitSuite) TestKMostRelevantMovies(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupMocks     func(r *resources, roomID model.RoomID, K int)
		roomID         model.RoomID
		K              int
		expectError    bool
		errorType      error
		errorContains  string
		expectedMovies []*model.MovieMeta
	}{

		{
			name: "Should return error when reducer panics",
			setupMocks: func(r *resources, roomID model.RoomID, K int) {
				prefs := []*model.Embedding{nil}
				r.embeddingRepository.On("LoadPreferenceEmbeddings", r.ctx, roomID).Return(prefs, nil).Once()
			},
			roomID:        validRoomID(),
			K:             2,
			expectError:   true,
			errorType:     ErrFailedToReduce,
			errorContains: "failed to reduce embeddings",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r, tc.roomID, tc.K)
			_, err := r.usecase.KMostRelevantMovies(r.ctx, tc.roomID, tc.K)
			assert.ErrorIs(t, err, tc.errorType)

		})
	}
}

func TestUnitSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseMovieUnitSuite))
}
