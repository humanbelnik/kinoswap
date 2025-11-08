// //go:build unit

package infra_postgres_vote

// import (
// 	"context"
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

// type VoteInfraUnitSuite struct {
// 	suite.Suite
// }

// type resources struct {
// 	db     *sqlx.DB
// 	mock   sqlmock.Sqlmock
// 	driver *Driver
// 	ctx    context.Context
// }

// func initResources(t provider.T) *resources {
// 	db, mock, err := sqlmock.New()
// 	if err != nil {
// 		t.Fatalf("failed to create sqlmock: %v", err)
// 	}

// 	sqlxDB := sqlx.NewDb(db, "sqlmock")
// 	driver := New(sqlxDB)

// 	return &resources{
// 		db:     sqlxDB,
// 		mock:   mock,
// 		driver: driver,
// 		ctx:    context.Background(),
// 	}
// }

// func validRoomID() model.RoomID {
// 	return model.RoomID("test-room-id")
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

// func (b *MovieMetaBuilder) Build() model.MovieMeta {
// 	return b.mm
// }

// func validVoteResult() model.VoteResult {
// 	movieMeta := NewMovieMetaBuilder().Build()
// 	return model.VoteResult{
// 		Results: map[*model.MovieMeta]model.Reaction{
// 			&movieMeta: model.PassReaction,
// 		},
// 	}
// }

// func validResultRow() struct {
// 	id         string
// 	title      string
// 	year       int
// 	rating     float64
// 	genres     pq.StringArray
// 	overview   string
// 	posterLink string
// 	passCount  int
// } {
// 	return struct {
// 		id         string
// 		title      string
// 		year       int
// 		rating     float64
// 		genres     pq.StringArray
// 		overview   string
// 		posterLink string
// 		passCount  int
// 	}{
// 		id:         uuid.New().String(),
// 		title:      "Test Movie",
// 		year:       2024,
// 		rating:     8.5,
// 		genres:     pq.StringArray{"Drama", "Comedy"},
// 		overview:   "Test overview",
// 		posterLink: "http://example.com/poster.jpg",
// 		passCount:  5,
// 	}
// }

// func (suite *VoteInfraUnitSuite) TestAddVote(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, roomID model.RoomID, voteResult model.VoteResult)
// 		roomID        model.RoomID
// 		voteResult    model.VoteResult
// 		expectError   bool
// 		errorContains string
// 	}{
// 		{
// 			name: "Should add vote successfully",
// 			setupMocks: func(r *resources, roomID model.RoomID, voteResult model.VoteResult) {
// 				r.mock.ExpectBegin()
// 				for movieMeta := range voteResult.Results {
// 					r.mock.ExpectExec("INSERT INTO results").
// 						WithArgs(string(roomID), movieMeta.ID).
// 						WillReturnResult(sqlmock.NewResult(1, 1))
// 				}
// 				r.mock.ExpectCommit()
// 			},
// 			roomID:      validRoomID(),
// 			voteResult:  validVoteResult(),
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return error when transaction fails",
// 			setupMocks: func(r *resources, roomID model.RoomID, voteResult model.VoteResult) {
// 				r.mock.ExpectBegin().WillReturnError(errors.New("transaction error"))
// 			},
// 			roomID:        validRoomID(),
// 			voteResult:    validVoteResult(),
// 			expectError:   true,
// 			errorContains: "failed to begin transaction",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.roomID, tc.voteResult)

// 			err := r.driver.AddVote(r.ctx, tc.roomID, tc.voteResult)

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

// func (suite *VoteInfraUnitSuite) TestLoadResults(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, roomID model.RoomID)
// 		roomID        model.RoomID
// 		expectError   bool
// 		errorContains string
// 		expectedCount int
// 	}{
// 		{
// 			name: "Should load results successfully",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				row := validResultRow()
// 				rows := sqlmock.NewRows([]string{
// 					"id", "title", "year", "rating", "genres", "overview", "poster_link", "pass_count",
// 				}).AddRow(
// 					row.id, row.title, row.year, row.rating, row.genres, row.overview, row.posterLink, row.passCount,
// 				)
// 				r.mock.ExpectQuery("SELECT").
// 					WithArgs(string(roomID)).
// 					WillReturnRows(rows)
// 			},
// 			roomID:        validRoomID(),
// 			expectError:   false,
// 			expectedCount: 1,
// 		},
// 		{
// 			name: "Should return error when query fails",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				r.mock.ExpectQuery("SELECT").
// 					WithArgs(string(roomID)).
// 					WillReturnError(errors.New("query error"))
// 			},
// 			roomID:        validRoomID(),
// 			expectError:   true,
// 			errorContains: "failed to load results",
// 		},
// 		{
// 			name: "Should return error when UUID is invalid",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				rows := sqlmock.NewRows([]string{
// 					"id", "title", "year", "rating", "genres", "overview", "poster_link", "pass_count",
// 				}).AddRow(
// 					"invalid-uuid", "Test Movie", 2024, 8.5, pq.StringArray{"Drama"}, "Overview", "poster.jpg", 5,
// 				)
// 				r.mock.ExpectQuery("SELECT").
// 					WithArgs(string(roomID)).
// 					WillReturnRows(rows)
// 			},
// 			roomID:        validRoomID(),
// 			expectError:   true,
// 			errorContains: "invalid movie UUID",
// 		},
// 	}

// 	for _, tc := range testCases {
// 		tc := tc
// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.roomID)

// 			results, err := r.driver.LoadResults(r.ctx, tc.roomID)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.ErrorContains(t, err, tc.errorContains)
// 				assert.Nil(t, results)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Len(t, results, tc.expectedCount)
// 				if tc.expectedCount > 0 {
// 					assert.Equal(t, tc.expectedCount, len(results))
// 					result := results[0]
// 					assert.Equal(t, "Test Movie", result.Title)
// 					assert.Equal(t, 2024, result.Year)
// 					assert.Equal(t, 8.5, result.Rating)
// 					assert.Equal(t, []string{"Drama", "Comedy"}, result.Genres)
// 				}
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func TestUnitSuite(t *testing.T) {
// 	suite.RunSuite(t, new(VoteInfraUnitSuite))
// }
