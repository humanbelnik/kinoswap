package infra_postgres_movie

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"testing"

// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/google/uuid"
// 	"github.com/humanbelnik/kinoswap/core/internal/model"
// 	"github.com/jmoiron/sqlx"
// 	"github.com/lib/pq"
// 	"github.com/ozontech/allure-go/pkg/framework/provider"
// 	"github.com/ozontech/allure-go/pkg/framework/suite"
// 	"github.com/stretchr/testify/assert"
// )

// type MovieInfraUnitSuite struct {
// 	suite.Suite
// }

// type resources struct {
// 	db         *sqlx.DB
// 	mock       sqlmock.Sqlmock
// 	repository *Repository
// 	ctx        context.Context
// }

// func initResources(t provider.T) *resources {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("failed to create sqlmock: %v", err)
// 	}

// 	sqlxDB := sqlx.NewDb(db, "sqlmock")
// 	repository := New(sqlxDB)

// 	return &resources{
// 		db:         sqlxDB,
// 		mock:       mock,
// 		repository: repository,
// 		ctx:        context.Background(),
// 	}
// }

// type MovieMetaBuilder struct {
// 	mm model.MovieMeta
// }

// func NewMovieMetaBuilder() *MovieMetaBuilder {
// 	return &MovieMetaBuilder{
// 		mm: model.MovieMeta{
// 			ID:         uuid.New(),
// 			PosterLink: "http://example.com/poster.jpg",
// 			Title:      "Test Movie",
// 			Genres:     []string{"Drama", "Comedy"},
// 			Year:       2024,
// 			Rating:     8.5,
// 			Overview:   "Test overview",
// 		},
// 	}
// }

// func (b *MovieMetaBuilder) WithID(id uuid.UUID) *MovieMetaBuilder {
// 	b.mm.ID = id
// 	return b
// }

// func (b *MovieMetaBuilder) WithTitle(title string) *MovieMetaBuilder {
// 	b.mm.Title = title
// 	return b
// }

// func (b *MovieMetaBuilder) Build() model.MovieMeta {
// 	return b.mm
// }

// func validEmbedding() model.Embedding {
// 	e := make(model.Embedding, 384)
// 	for i := range e {
// 		e[i] = float32(i) * 0.01
// 	}
// 	return e
// }

// func (suite *MovieInfraUnitSuite) TestStore(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, movieMeta model.MovieMeta)
// 		movieMeta     model.MovieMeta
// 		expectError   bool
// 		errorContains string
// 	}{
// 		{
// 			name: "Should store movie successfully",
// 			setupMocks: func(r *resources, movieMeta model.MovieMeta) {
// 				r.mock.ExpectExec("INSERT INTO movies").
// 					WithArgs(
// 						movieMeta.ID,
// 						movieMeta.Title,
// 						movieMeta.Year,
// 						movieMeta.Rating,
// 						pq.StringArray(movieMeta.Genres),
// 						movieMeta.Overview,
// 						movieMeta.PosterLink,
// 					).
// 					WillReturnResult(sqlmock.NewResult(1, 1))
// 			},
// 			movieMeta:   NewMovieMetaBuilder().Build(),
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return error when insert fails",
// 			setupMocks: func(r *resources, movieMeta model.MovieMeta) {
// 				r.mock.ExpectExec("INSERT INTO movies").
// 					WithArgs(
// 						movieMeta.ID,
// 						movieMeta.Title,
// 						movieMeta.Year,
// 						movieMeta.Rating,
// 						pq.StringArray(movieMeta.Genres),
// 						movieMeta.Overview,
// 						movieMeta.PosterLink,
// 					).
// 					WillReturnError(errors.New("insert error"))
// 			},
// 			movieMeta:     NewMovieMetaBuilder().Build(),
// 			expectError:   true,
// 			errorContains: "failed to store movie",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.movieMeta)

// 			err := r.repository.Store(r.ctx, tc.movieMeta)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.ErrorContains(t, err, tc.errorContains)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func (suite *MovieInfraUnitSuite) TestLoad(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources)
// 		expectError   bool
// 		errorContains string
// 		expectedCount int
// 	}{
// 		{
// 			name: "Should load movies successfully",
// 			setupMocks: func(r *resources) {
// 				movie1 := NewMovieMetaBuilder().Build()
// 				movie2 := NewMovieMetaBuilder().Build()
// 				rows := sqlmock.NewRows([]string{
// 					"id", "title", "year", "rating", "genres", "overview", "poster_link",
// 				}).
// 					AddRow(
// 						movie1.ID, movie1.Title, movie1.Year, movie1.Rating,
// 						pq.StringArray(movie1.Genres), movie1.Overview, movie1.PosterLink,
// 					).
// 					AddRow(
// 						movie2.ID, movie2.Title, movie2.Year, movie2.Rating,
// 						pq.StringArray(movie2.Genres), movie2.Overview, movie2.PosterLink,
// 					)
// 				r.mock.ExpectQuery("SELECT id, title, year, rating, genres, overview, poster_link FROM movies").
// 					WillReturnRows(rows)
// 			},
// 			expectError:   false,
// 			expectedCount: 2,
// 		},
// 		{
// 			name: "Should return error when query fails",
// 			setupMocks: func(r *resources) {
// 				r.mock.ExpectQuery("SELECT id, title, year, rating, genres, overview, poster_link FROM movies").
// 					WillReturnError(errors.New("query error"))
// 			},
// 			expectError:   true,
// 			errorContains: "failed to query movies",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r)

// 			movies, err := r.repository.Load(r.ctx)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.ErrorContains(t, err, tc.errorContains)
// 				assert.Nil(t, movies)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Len(t, movies, tc.expectedCount)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func (suite *MovieInfraUnitSuite) TestLoadByID(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, movieID uuid.UUID)
// 		movieID       uuid.UUID
// 		expectError   bool
// 		errorType     error
// 		errorContains string
// 	}{
// 		{
// 			name: "Should load movie by ID successfully",
// 			setupMocks: func(r *resources, movieID uuid.UUID) {
// 				movie := NewMovieMetaBuilder().WithID(movieID).Build()
// 				rows := sqlmock.NewRows([]string{
// 					"id", "title", "year", "rating", "genres", "overview", "poster_link",
// 				}).
// 					AddRow(
// 						movie.ID, movie.Title, movie.Year, movie.Rating,
// 						pq.StringArray(movie.Genres), movie.Overview, movie.PosterLink,
// 					)
// 				r.mock.ExpectQuery("SELECT id, title, year, rating, genres, overview, poster_link FROM movies WHERE id = ?").
// 					WithArgs(movieID).
// 					WillReturnRows(rows)
// 			},
// 			movieID:     uuid.New(),
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return ErrMovieNotFound when movie not found",
// 			setupMocks: func(r *resources, movieID uuid.UUID) {
// 				r.mock.ExpectQuery("SELECT id, title, year, rating, genres, overview, poster_link FROM movies WHERE id = ?").
// 					WithArgs(movieID).
// 					WillReturnError(sql.ErrNoRows)
// 			},
// 			movieID:       uuid.New(),
// 			expectError:   true,
// 			errorType:     ErrMovieNotFound,
// 			errorContains: "movie not found",
// 		},
// 		{
// 			name: "Should return error when query fails",
// 			setupMocks: func(r *resources, movieID uuid.UUID) {
// 				r.mock.ExpectQuery("SELECT id, title, year, rating, genres, overview, poster_link FROM movies WHERE id = ?").
// 					WithArgs(movieID).
// 					WillReturnError(errors.New("query error"))
// 			},
// 			movieID:       uuid.New(),
// 			expectError:   true,
// 			errorContains: "failed to load movie by id",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.movieID)

// 			movie, err := r.repository.LoadByID(r.ctx, tc.movieID)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				if tc.errorType != nil {
// 					assert.ErrorIs(t, err, tc.errorType)
// 				}
// 				assert.ErrorContains(t, err, tc.errorContains)
// 				assert.Equal(t, model.MovieMeta{}, movie)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tc.movieID, movie.ID)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func (suite *MovieInfraUnitSuite) TestLoadByIDs(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, movieIDs []uuid.UUID)
// 		movieIDs      []uuid.UUID
// 		expectError   bool
// 		errorContains string
// 		expectedCount int
// 	}{
// 		{
// 			name: "Should load movies by IDs successfully",
// 			setupMocks: func(r *resources, movieIDs []uuid.UUID) {
// 				movie1 := NewMovieMetaBuilder().WithID(movieIDs[0]).Build()
// 				movie2 := NewMovieMetaBuilder().WithID(movieIDs[1]).Build()
// 				rows := sqlmock.NewRows([]string{
// 					"id", "title", "year", "rating", "genres", "overview", "poster_link",
// 				}).
// 					AddRow(
// 						movie1.ID, movie1.Title, movie1.Year, movie1.Rating,
// 						pq.StringArray(movie1.Genres), movie1.Overview, movie1.PosterLink,
// 					).
// 					AddRow(
// 						movie2.ID, movie2.Title, movie2.Year, movie2.Rating,
// 						pq.StringArray(movie2.Genres), movie2.Overview, movie2.PosterLink,
// 					)
// 				r.mock.ExpectQuery("SELECT id, title, year, rating, genres, overview, poster_link FROM movies WHERE id IN").
// 					WithArgs(movieIDs[0], movieIDs[1]).
// 					WillReturnRows(rows)
// 			},
// 			movieIDs:      []uuid.UUID{uuid.New(), uuid.New()},
// 			expectError:   false,
// 			expectedCount: 2,
// 		},
// 		{
// 			name: "Should return empty slice when no IDs provided",
// 			setupMocks: func(r *resources, movieIDs []uuid.UUID) {
// 				// No query expected for empty IDs
// 			},
// 			movieIDs:      []uuid.UUID{},
// 			expectError:   false,
// 			expectedCount: 0,
// 		},
// 		{
// 			name: "Should return error when query fails",
// 			setupMocks: func(r *resources, movieIDs []uuid.UUID) {
// 				r.mock.ExpectQuery("SELECT id, title, year, rating, genres, overview, poster_link FROM movies WHERE id IN").
// 					WithArgs(movieIDs[0], movieIDs[1]).
// 					WillReturnError(errors.New("query error"))
// 			},
// 			movieIDs:      []uuid.UUID{uuid.New(), uuid.New()},
// 			expectError:   true,
// 			errorContains: "failed to query movies by ids",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.movieIDs)

// 			movies, err := r.repository.LoadByIDs(r.ctx, tc.movieIDs)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.ErrorContains(t, err, tc.errorContains)
// 				assert.Nil(t, movies)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Len(t, movies, tc.expectedCount)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func (suite *MovieInfraUnitSuite) TestUpdate(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, movieMeta model.MovieMeta)
// 		movieMeta     model.MovieMeta
// 		expectError   bool
// 		errorType     error
// 		errorContains string
// 	}{
// 		{
// 			name: "Should update movie successfully",
// 			setupMocks: func(r *resources, movieMeta model.MovieMeta) {
// 				r.mock.ExpectExec("UPDATE movies").
// 					WithArgs(
// 						movieMeta.Title,
// 						movieMeta.Year,
// 						movieMeta.Rating,
// 						pq.StringArray(movieMeta.Genres),
// 						movieMeta.Overview,
// 						movieMeta.PosterLink,
// 						movieMeta.ID,
// 					).
// 					WillReturnResult(sqlmock.NewResult(0, 1))
// 			},
// 			movieMeta:   NewMovieMetaBuilder().Build(),
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return ErrMovieNotFound when movie not found",
// 			setupMocks: func(r *resources, movieMeta model.MovieMeta) {
// 				r.mock.ExpectExec("UPDATE movies").
// 					WithArgs(
// 						movieMeta.Title,
// 						movieMeta.Year,
// 						movieMeta.Rating,
// 						pq.StringArray(movieMeta.Genres),
// 						movieMeta.Overview,
// 						movieMeta.PosterLink,
// 						movieMeta.ID,
// 					).
// 					WillReturnResult(sqlmock.NewResult(0, 0))
// 			},
// 			movieMeta:     NewMovieMetaBuilder().Build(),
// 			expectError:   true,
// 			errorType:     ErrMovieNotFound,
// 			errorContains: "movie not found",
// 		},
// 		{
// 			name: "Should return error when update fails",
// 			setupMocks: func(r *resources, movieMeta model.MovieMeta) {
// 				r.mock.ExpectExec("UPDATE movies").
// 					WithArgs(
// 						movieMeta.Title,
// 						movieMeta.Year,
// 						movieMeta.Rating,
// 						pq.StringArray(movieMeta.Genres),
// 						movieMeta.Overview,
// 						movieMeta.PosterLink,
// 						movieMeta.ID,
// 					).
// 					WillReturnError(errors.New("update error"))
// 			},
// 			movieMeta:     NewMovieMetaBuilder().Build(),
// 			expectError:   true,
// 			errorContains: "failed to update movie",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.movieMeta)

// 			err := r.repository.Update(r.ctx, tc.movieMeta)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				if tc.errorType != nil {
// 					assert.ErrorIs(t, err, tc.errorType)
// 				}
// 				assert.ErrorContains(t, err, tc.errorContains)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func (suite *MovieInfraUnitSuite) TestDeleteByID(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, movieID uuid.UUID)
// 		movieID       uuid.UUID
// 		expectError   bool
// 		errorType     error
// 		errorContains string
// 	}{
// 		{
// 			name: "Should delete movie successfully",
// 			setupMocks: func(r *resources, movieID uuid.UUID) {
// 				r.mock.ExpectExec("DELETE FROM movies WHERE id = ?").
// 					WithArgs(movieID).
// 					WillReturnResult(sqlmock.NewResult(0, 1))
// 			},
// 			movieID:     uuid.New(),
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return ErrMovieNotFound when movie not found",
// 			setupMocks: func(r *resources, movieID uuid.UUID) {
// 				r.mock.ExpectExec("DELETE FROM movies WHERE id = ?").
// 					WithArgs(movieID).
// 					WillReturnResult(sqlmock.NewResult(0, 0))
// 			},
// 			movieID:       uuid.New(),
// 			expectError:   true,
// 			errorType:     ErrMovieNotFound,
// 			errorContains: "movie not found",
// 		},
// 		{
// 			name: "Should return error when delete fails",
// 			setupMocks: func(r *resources, movieID uuid.UUID) {
// 				r.mock.ExpectExec("DELETE FROM movies WHERE id = ?").
// 					WithArgs(movieID).
// 					WillReturnError(errors.New("delete error"))
// 			},
// 			movieID:       uuid.New(),
// 			expectError:   true,
// 			errorContains: "failed to delete movie",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.movieID)

// 			err := r.repository.DeleteByID(r.ctx, tc.movieID)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				if tc.errorType != nil {
// 					assert.ErrorIs(t, err, tc.errorType)
// 				}
// 				assert.ErrorContains(t, err, tc.errorContains)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func (suite *MovieInfraUnitSuite) TestUpdateEmbedding(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, movieID uuid.UUID, embedding model.Embedding)
// 		movieID       uuid.UUID
// 		embedding     model.Embedding
// 		expectError   bool
// 		errorType     error
// 		errorContains string
// 	}{

// 		{
// 			name: "Should return ErrInvalidEmbedding when embedding has wrong dimensions",
// 			setupMocks: func(r *resources, movieID uuid.UUID, embedding model.Embedding) {
// 				// No query expected for invalid embedding
// 			},
// 			movieID:       uuid.New(),
// 			embedding:     make(model.Embedding, 100), // Wrong dimension
// 			expectError:   true,
// 			errorType:     ErrInvalidEmbedding,
// 			errorContains: "invalid embedding dimensions",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.movieID, tc.embedding)

// 			err := r.repository.UpdateEmbedding(r.ctx, tc.movieID, tc.embedding)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				if tc.errorType != nil {
// 					assert.ErrorIs(t, err, tc.errorType)
// 				}
// 				assert.ErrorContains(t, err, tc.errorContains)
// 			} else {
// 				assert.NoError(t, err)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func (suite *MovieInfraUnitSuite) TestKNN(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, k int, embedding model.Embedding)
// 		k             int
// 		embedding     model.Embedding
// 		expectError   bool
// 		errorType     error
// 		errorContains string
// 		expectedCount int
// 	}{
// 		{
// 			name: "Should return ErrInvalidEmbedding when embedding has wrong dimensions",
// 			setupMocks: func(r *resources, k int, embedding model.Embedding) {
// 				// No query expected for invalid embedding
// 			},
// 			k:             2,
// 			embedding:     make(model.Embedding, 100), // Wrong dimension
// 			expectError:   true,
// 			errorType:     ErrInvalidEmbedding,
// 			errorContains: "invalid embedding dimensions",
// 		},
// 		{
// 			name: "Should return error when k is not positive",
// 			setupMocks: func(r *resources, k int, embedding model.Embedding) {
// 				// No query expected for invalid k
// 			},
// 			k:             0,
// 			embedding:     validEmbedding(),
// 			expectError:   true,
// 			errorContains: "k must be positive",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.k, tc.embedding)

// 			movies, err := r.repository.KNN(r.ctx, tc.k, tc.embedding)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				if tc.errorType != nil {
// 					assert.ErrorIs(t, err, tc.errorType)
// 				}
// 				assert.ErrorContains(t, err, tc.errorContains)
// 				assert.Nil(t, movies)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Len(t, movies, tc.expectedCount)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func TestUnitSuite(t *testing.T) {
// 	suite.RunSuite(t, new(MovieInfraUnitSuite))
// }
