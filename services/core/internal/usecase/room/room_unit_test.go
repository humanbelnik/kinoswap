package usecase_room

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	embedder_mocks "github.com/humanbelnik/kinoswap/core/internal/usecase/room/mocks/room/Embedder"
	repo_mocks "github.com/humanbelnik/kinoswap/core/internal/usecase/room/mocks/room/repository"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseRoomUnitSuite struct {
	suite.Suite
}

type resources struct {
	usecase  *Usecase
	roomRepo *repo_mocks.RoomRepository
	embedder *embedder_mocks.Embedder
	ctx      context.Context
}

func initResources(t provider.T) *resources {
	roomRepo := repo_mocks.NewRoomRepository(t)
	embedder := embedder_mocks.NewEmbedder(t)
	usecase := New(roomRepo, embedder, 20)

	return &resources{
		roomRepo: roomRepo,
		embedder: embedder,
		usecase:  usecase,
		ctx:      context.Background(),
	}
}

func validRoomCode() string {
	return "123456"
}

func validOwnerID() string {
	return uuid.New().String()
}

func (suite *UsecaseRoomUnitSuite) TestBook(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources)
		expectError   bool
		expectedError error
	}{
		{
			name: "Should book room successfully",
			setupMocks: func(r *resources) {
				// createRoomLobby будет вызывать CreateAndBook 1 раз успешно
				r.roomRepo.On("CreateAndBook", r.ctx, mock.AnythingOfType("model.Room"), mock.AnythingOfType("uuid.UUID")).
					Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources) {
				// createRoomLobby будет пытаться 3 раза и все вернут ошибку
				r.roomRepo.On("CreateAndBook", r.ctx, mock.AnythingOfType("model.Room"), mock.AnythingOfType("uuid.UUID")).
					Return(ErrCodeConflict).Times(3)
			},
			expectError:   true,
			expectedError: ErrRoomsUnavailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r)

			roomCode, ownerToken, err := r.usecase.Book(r.ctx)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.expectedError)
				assert.Empty(t, roomCode)
				assert.Empty(t, ownerToken)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, roomCode)
				assert.NotEmpty(t, ownerToken)
			}
			r.roomRepo.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseRoomUnitSuite) TestIsOwner(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, code string, ownerID string)
		expected      bool
		expectError   bool
		expectedError error
	}{
		{
			name: "Should return true when user is owner",
			setupMocks: func(r *resources, code string, ownerID string) {
				ownerUUID, _ := uuid.Parse(ownerID)
				r.roomRepo.On("IsOwner", r.ctx, code, ownerUUID).Return(true, nil).Once()
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string, ownerID string) {
				ownerUUID, _ := uuid.Parse(ownerID)
				r.roomRepo.On("IsOwner", r.ctx, code, ownerUUID).Return(false, ErrResourceNotFound).Once()
			},
			expected:      false,
			expectError:   true,
			expectedError: ErrResourceNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validRoomCode()
			ownerID := validOwnerID()
			tc.setupMocks(r, code, ownerID)

			result, err := r.usecase.IsOwner(r.ctx, code, ownerID)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected, result)
			r.roomRepo.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseRoomUnitSuite) TestFree(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, code string)
		expectError   bool
		expectedError error
	}{
		{
			name: "Should free room successfully",
			setupMocks: func(r *resources, code string) {
				r.roomRepo.On("DeleteByCode", r.ctx, code).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string) {
				r.roomRepo.On("DeleteByCode", r.ctx, code).Return(ErrResourceNotFound).Once()
			},
			expectError:   true,
			expectedError: ErrResourceNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validRoomCode()
			tc.setupMocks(r, code)

			err := r.usecase.Free(r.ctx, code)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			r.roomRepo.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseRoomUnitSuite) TestStatus(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, code string)
		expected      string
		expectError   bool
		expectedError error
	}{
		{
			name: "Should return status successfully",
			setupMocks: func(r *resources, code string) {
				r.roomRepo.On("StatusByCode", r.ctx, code).Return("LOBBY", nil).Once()
			},
			expected:    "LOBBY",
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string) {
				r.roomRepo.On("StatusByCode", r.ctx, code).Return("", ErrResourceNotFound).Once()
			},
			expected:      "",
			expectError:   true,
			expectedError: ErrResourceNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validRoomCode()
			tc.setupMocks(r, code)

			result, err := r.usecase.Status(r.ctx, code)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected, result)
			r.roomRepo.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseRoomUnitSuite) TestSetStatus(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, code string, status string)
		expectError   bool
		expectedError error
	}{
		{
			name: "Should set status successfully",
			setupMocks: func(r *resources, code string, status string) {
				r.roomRepo.On("SetStatusByCode", r.ctx, code, status).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string, status string) {
				r.roomRepo.On("SetStatusByCode", r.ctx, code, status).Return(ErrResourceNotFound).Once()
			},
			expectError:   true,
			expectedError: ErrResourceNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validRoomCode()
			status := "VOTING"
			tc.setupMocks(r, code, status)

			err := r.usecase.SetStatus(r.ctx, code, status)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			r.roomRepo.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseRoomUnitSuite) TestParticipate(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, code string, pref model.Preference, userID *string)
		expectError   bool
		expectedError error
	}{
		{
			name: "Should participate successfully with new userID",
			setupMocks: func(r *resources, code string, pref model.Preference, userID *string) {
				r.embedder.On("BuildPreferenceEmbedding", r.ctx, pref).Return(model.Embedding(make([]float32, model.EmbeddingDimension)), nil).Once()
				r.roomRepo.On("AddPreferenceEmbedding", r.ctx, code, mock.AnythingOfType("uuid.UUID"), model.Embedding(make([]float32, model.EmbeddingDimension))).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string, pref model.Preference, userID *string) {
				r.embedder.On("BuildPreferenceEmbedding", r.ctx, pref).Return(model.Embedding(make([]float32, model.EmbeddingDimension)), nil).Once()
				r.roomRepo.On("AddPreferenceEmbedding", r.ctx, code, mock.AnythingOfType("uuid.UUID"), model.Embedding(make([]float32, model.EmbeddingDimension))).Return(ErrResourceNotFound).Once()
			},
			expectError:   true,
			expectedError: ErrResourceNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validRoomCode()
			pref := model.Preference{Text: "test preference"}
			var userID *string = nil // new user
			tc.setupMocks(r, code, pref, userID)

			returnedUserID, err := r.usecase.Participate(r.ctx, code, pref, userID)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, returnedUserID)
			}
			r.embedder.AssertExpectations(t)
			r.roomRepo.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseRoomUnitSuite) TestParticipantsCount(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, code string)
		expected      int
		expectError   bool
		expectedError error
	}{
		{
			name: "Should return participants count successfully",
			setupMocks: func(r *resources, code string) {
				r.roomRepo.On("ParticipantsCount", r.ctx, code).Return(5, nil).Once()
			},
			expected:    5,
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string) {
				r.roomRepo.On("ParticipantsCount", r.ctx, code).Return(0, ErrResourceNotFound).Once()
			},
			expected:      0,
			expectError:   true,
			expectedError: ErrResourceNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validRoomCode()
			tc.setupMocks(r, code)

			result, err := r.usecase.ParticipantsCount(r.ctx, code)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected, result)
			r.roomRepo.AssertExpectations(t)
		})
	}
}

func TestUnitSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseRoomUnitSuite))
}
