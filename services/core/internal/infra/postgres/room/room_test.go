// //go:build unit

package infra_postgres_room

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"testing"

// 	"github.com/DATA-DOG/go-sqlmock"
// 	"github.com/humanbelnik/kinoswap/core/internal/model"
// 	"github.com/jmoiron/sqlx"
// 	"github.com/ozontech/allure-go/pkg/framework/provider"
// 	"github.com/ozontech/allure-go/pkg/framework/suite"
// 	"github.com/stretchr/testify/assert"
// )

// type RoomInfraUnitSuite struct {
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

// type PreferenceBuilder struct {
// 	p model.Preference
// }

// func NewPreferenceBuilder() *PreferenceBuilder {
// 	return &PreferenceBuilder{
// 		p: model.Preference{
// 			Text: "test preference text",
// 		},
// 	}
// }

// func (b *PreferenceBuilder) WithText(text string) *PreferenceBuilder {
// 	b.p.Text = text
// 	return b
// }

// func (b *PreferenceBuilder) Build() model.Preference {
// 	return b.p
// }

// func (suite *RoomInfraUnitSuite) TestCreateAndAquire(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, roomID model.RoomID)
// 		roomID        model.RoomID
// 		expectError   bool
// 		errorContains string
// 	}{
// 		{
// 			name: "Should create and acquire room successfully",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				r.mock.ExpectExec("INSERT INTO rooms").
// 					WithArgs(string(roomID), RoomStatusAcquired).
// 					WillReturnResult(sqlmock.NewResult(1, 1))
// 			},
// 			roomID:      validRoomID(),
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return error when insert fails",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				r.mock.ExpectExec("INSERT INTO rooms").
// 					WithArgs(string(roomID), RoomStatusAcquired).
// 					WillReturnError(errors.New("insert error"))
// 			},
// 			roomID:        validRoomID(),
// 			expectError:   true,
// 			errorContains: "insert error",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.roomID)

// 			err := r.driver.CreateAndAquire(r.ctx, tc.roomID)

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

// func (suite *RoomInfraUnitSuite) TestFindAndAcquire(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources)
// 		expected      model.RoomID
// 		expectError   bool
// 		errorContains string
// 	}{
// 		{
// 			name: "Should find and acquire room successfully",
// 			setupMocks: func(r *resources) {
// 				rows := sqlmock.NewRows([]string{"id"}).AddRow("test-room-id")
// 				r.mock.ExpectQuery("UPDATE rooms").
// 					WithArgs(RoomStatusAcquired, RoomStatusFree).
// 					WillReturnRows(rows)
// 			},
// 			expected:    model.RoomID("test-room-id"),
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return error when no room found",
// 			setupMocks: func(r *resources) {
// 				r.mock.ExpectQuery("UPDATE rooms").
// 					WithArgs(RoomStatusAcquired, RoomStatusFree).
// 					WillReturnError(sql.ErrNoRows)
// 			},
// 			expected:      model.EmptyRoomID,
// 			expectError:   true,
// 			errorContains: "no rows in result set",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r)

// 			result, err := r.driver.FindAndAcquire(r.ctx)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.ErrorContains(t, err, tc.errorContains)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tc.expected, result)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func (suite *RoomInfraUnitSuite) TestTryAcquire(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, roomID model.RoomID)
// 		roomID        model.RoomID
// 		expectError   bool
// 		errorContains string
// 	}{
// 		{
// 			name: "Should try acquire room successfully",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				r.mock.ExpectExec("UPDATE rooms SET status").
// 					WithArgs(RoomStatusAcquired, string(roomID), RoomStatusFree).
// 					WillReturnResult(sqlmock.NewResult(0, 1))
// 			},
// 			roomID:      validRoomID(),
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return error when no rows affected",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				r.mock.ExpectExec("UPDATE rooms SET status").
// 					WithArgs(RoomStatusAcquired, string(roomID), RoomStatusFree).
// 					WillReturnResult(sqlmock.NewResult(0, 0))
// 			},
// 			roomID:        validRoomID(),
// 			expectError:   true,
// 			errorContains: "failed to acquire",
// 		},
// 		{
// 			name: "Should return error when update fails",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				r.mock.ExpectExec("UPDATE rooms SET status").
// 					WithArgs(RoomStatusAcquired, string(roomID), RoomStatusFree).
// 					WillReturnError(errors.New("update error"))
// 			},
// 			roomID:        validRoomID(),
// 			expectError:   true,
// 			errorContains: "update error",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.roomID)

// 			err := r.driver.TryAcquire(r.ctx, tc.roomID)

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

// func (suite *RoomInfraUnitSuite) TestIsExistsRoomID(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, roomID model.RoomID)
// 		roomID        model.RoomID
// 		expected      bool
// 		expectError   bool
// 		errorContains string
// 	}{
// 		{
// 			name: "Should return true when room exists",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
// 				r.mock.ExpectQuery("SELECT EXISTS").
// 					WithArgs(string(roomID)).
// 					WillReturnRows(rows)
// 			},
// 			roomID:      validRoomID(),
// 			expected:    true,
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return false when room does not exist",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
// 				r.mock.ExpectQuery("SELECT EXISTS").
// 					WithArgs(string(roomID)).
// 					WillReturnRows(rows)
// 			},
// 			roomID:      validRoomID(),
// 			expected:    false,
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return error when query fails",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				r.mock.ExpectQuery("SELECT EXISTS").
// 					WithArgs(string(roomID)).
// 					WillReturnError(errors.New("query error"))
// 			},
// 			roomID:        validRoomID(),
// 			expected:      false,
// 			expectError:   true,
// 			errorContains: "query error",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.roomID)

// 			result, err := r.driver.IsExistsRoomID(r.ctx, tc.roomID)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.ErrorContains(t, err, tc.errorContains)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tc.expected, result)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func (suite *RoomInfraUnitSuite) TestAppendPreference(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, roomID model.RoomID, preference model.Preference)
// 		roomID        model.RoomID
// 		preference    model.Preference
// 		expectError   bool
// 		errorContains string
// 	}{
// 		{
// 			name: "Should append preference successfully",
// 			setupMocks: func(r *resources, roomID model.RoomID, preference model.Preference) {
// 				r.mock.ExpectExec("UPDATE rooms").
// 					WithArgs(preference.Text, string(roomID), RoomStatusAcquired).
// 					WillReturnResult(sqlmock.NewResult(0, 1))
// 			},
// 			roomID:      validRoomID(),
// 			preference:  NewPreferenceBuilder().Build(),
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return error when update fails",
// 			setupMocks: func(r *resources, roomID model.RoomID, preference model.Preference) {
// 				r.mock.ExpectExec("UPDATE rooms").
// 					WithArgs(preference.Text, string(roomID), RoomStatusAcquired).
// 					WillReturnError(errors.New("update error"))
// 			},
// 			roomID:        validRoomID(),
// 			preference:    NewPreferenceBuilder().Build(),
// 			expectError:   true,
// 			errorContains: "update error",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.roomID, tc.preference)

// 			err := r.driver.AppendPreference(r.ctx, tc.roomID, tc.preference)

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

// func (suite *RoomInfraUnitSuite) TestIsRoomAcquired(t provider.T) {
// 	t.Parallel()

// 	testCases := []struct {
// 		name          string
// 		setupMocks    func(r *resources, roomID model.RoomID)
// 		roomID        model.RoomID
// 		expected      bool
// 		expectError   bool
// 		errorContains string
// 	}{
// 		{
// 			name: "Should return true when room is acquired",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
// 				r.mock.ExpectQuery("SELECT EXISTS").
// 					WithArgs(string(roomID), RoomStatusAcquired).
// 					WillReturnRows(rows)
// 			},
// 			roomID:      validRoomID(),
// 			expected:    true,
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return false when room is not acquired",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				rows := sqlmock.NewRows([]string{"exists"}).AddRow(false)
// 				r.mock.ExpectQuery("SELECT EXISTS").
// 					WithArgs(string(roomID), RoomStatusAcquired).
// 					WillReturnRows(rows)
// 			},
// 			roomID:      validRoomID(),
// 			expected:    false,
// 			expectError: false,
// 		},
// 		{
// 			name: "Should return error when query fails",
// 			setupMocks: func(r *resources, roomID model.RoomID) {
// 				r.mock.ExpectQuery("SELECT EXISTS").
// 					WithArgs(string(roomID), RoomStatusAcquired).
// 					WillReturnError(errors.New("query error"))
// 			},
// 			roomID:        validRoomID(),
// 			expected:      false,
// 			expectError:   true,
// 			errorContains: "query error",
// 		},
// 	}

// 	for _, tc := range testCases {

// 		t.Run(tc.name, func(t provider.T) {
// 			t.Parallel()
// 			r := initResources(t)
// 			tc.setupMocks(r, tc.roomID)

// 			result, err := r.driver.IsRoomAcquired(r.ctx, tc.roomID)

// 			if tc.expectError {
// 				assert.Error(t, err)
// 				assert.ErrorContains(t, err, tc.errorContains)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, tc.expected, result)
// 			}
// 			assert.NoError(t, r.mock.ExpectationsWereMet())
// 		})
// 	}
// }

// func TestUnitSuite(t *testing.T) {
// 	suite.RunSuite(t, new(RoomInfraUnitSuite))
// }
