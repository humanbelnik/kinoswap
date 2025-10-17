//go:build unit

package usecase_room

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/humanbelnik/kinoswap/core/internal/model"
	cache_mocks "github.com/humanbelnik/kinoswap/core/mocks/room/cache"
	embedder_mocks "github.com/humanbelnik/kinoswap/core/mocks/room/embedder"
	repo_mocks "github.com/humanbelnik/kinoswap/core/mocks/room/repository"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseRoomUnitSuite struct {
	suite.Suite
}

type resources struct {
	usecase       *Usecase
	roomRepo      *repo_mocks.RoomRepository
	embeddingRepo *repo_mocks.EmbeddingRepository
	set           *cache_mocks.EmptyRoomsIDSet
	embedder      *embedder_mocks.Embedder
	ctx           context.Context
	wg            sync.WaitGroup
}

func validRoomID() model.RoomID {
	return model.RoomID("123456")
}

func initResources(t provider.T) *resources {
	roomRepo := repo_mocks.NewRoomRepository(t)
	embeddingRepo := repo_mocks.NewEmbeddingRepository(t)
	set := cache_mocks.NewEmptyRoomsIDSet(t)
	embedder := embedder_mocks.NewEmbedder(t)
	usecase := New(roomRepo, embedder, embeddingRepo, set)

	return &resources{
		roomRepo:      roomRepo,
		embeddingRepo: embeddingRepo,
		set:           set,
		embedder:      embedder,
		usecase:       usecase,
		ctx:           context.Background(),
	}
}

func (suite *UsecaseRoomUnitSuite) TestIsRoomAcquired(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		setupMocks  func(r *resources, roomID model.RoomID)
		expected    bool
		expectError bool
		expectedErr error
	}{
		{
			name: "Should return true when repository returns true",
			setupMocks: func(r *resources, roomID model.RoomID) {
				r.roomRepo.On("IsRoomAcquired", r.ctx, roomID).Return(true, nil).Once()
			},
			expected:    true,
			expectError: false,
		},
		{
			name: "Should return false when repository returns false",
			setupMocks: func(r *resources, roomID model.RoomID) {
				r.roomRepo.On("IsRoomAcquired", r.ctx, roomID).Return(false, nil).Once()
			},
			expected:    false,
			expectError: false,
		},
		{
			name: "Should return false and error when repository returns error",
			setupMocks: func(r *resources, roomID model.RoomID) {
				expectedError := errors.New("repository error")
				r.roomRepo.On("IsRoomAcquired", r.ctx, roomID).Return(false, expectedError).Once()
			},
			expected:    false,
			expectError: true,
			expectedErr: errors.New("repository error"),
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			roomID := validRoomID()
			tc.setupMocks(r, roomID)

			result, err := r.usecase.IsRoomAcquired(r.ctx, roomID)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.expected, result)
			r.roomRepo.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseRoomUnitSuite) TestParticipate(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, roomID model.RoomID, preference model.Preference)
		expectError   bool
		errorType     error
		errorContains string
	}{
		{
			name: "Should participate successfully",
			setupMocks: func(r *resources, roomID model.RoomID, preference model.Preference) {
				r.roomRepo.On("AppendPreference", r.ctx, roomID, preference).Return(nil).Once()

				r.wg.Add(2)
				r.embedder.On("BuildPreferenceEmbedding", mock.Anything, preference).Return(model.Embedding{}, nil).Run(func(args mock.Arguments) {
					r.wg.Done()
				})
				r.embeddingRepo.On("Append", mock.Anything, roomID, mock.AnythingOfType("model.Embedding")).Return(nil).Run(func(args mock.Arguments) {
					r.wg.Done()
				})
			},
			expectError: false,
		},
		{
			name: "Should return ErrParticipate when repository fails",
			setupMocks: func(r *resources, roomID model.RoomID, preference model.Preference) {
				repoError := errors.New("repository error")
				r.roomRepo.On("AppendPreference", r.ctx, roomID, preference).Return(repoError).Once()
			},
			expectError:   true,
			errorType:     ErrParticipate,
			errorContains: "repository error",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			roomID := validRoomID()
			preference := model.Preference{Text: "test preference"}
			tc.setupMocks(r, roomID, preference)

			err := r.usecase.Participate(r.ctx, roomID, preference)
			r.wg.Wait()

			if tc.expectError {
				assert.ErrorIs(t, err, tc.errorType)
				assert.ErrorContains(t, err, tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
			r.roomRepo.AssertExpectations(t)
			if !tc.expectError {
				r.embedder.AssertExpectations(t)
				r.embeddingRepo.AssertExpectations(t)
			}
		})
	}
}

func (suite *UsecaseRoomUnitSuite) TestAcquireRoom(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupMocks     func(r *resources)
		expectedRoomID model.RoomID
		expectError    bool
		errorType      error
		errorContains  string
	}{
		{
			name: "Should acquire room successfully from cache",
			setupMocks: func(r *resources) {
				roomID := validRoomID()
				r.set.On("Remove", r.ctx).Return(roomID, nil).Once()
				r.roomRepo.On("TryAcquire", r.ctx, roomID).Return(nil).Once()
			},
			expectedRoomID: validRoomID(),
			expectError:    false,
		},
		{
			name: "Should return ErrCreateRoom when all paths fail",
			setupMocks: func(r *resources) {
				createError := errors.New("create error")
				r.set.On("Remove", r.ctx).Return(model.EmptyRoomID, nil).Once()
				r.roomRepo.On("FindAndAcquire", r.ctx).Return(model.EmptyRoomID, nil).Once()
				r.roomRepo.On("IsExistsRoomID", r.ctx, mock.AnythingOfType("model.RoomID")).Return(false, nil).Once()
				r.roomRepo.On("CreateAndAquire", r.ctx, mock.AnythingOfType("model.RoomID")).Return(createError).Once()
			},
			expectedRoomID: model.EmptyRoomID,
			expectError:    true,
			errorType:      ErrCreateRoom,
			errorContains:  "create error",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			tc.setupMocks(r)

			result, err := r.usecase.AcquireRoom(r.ctx)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.errorType)
				assert.ErrorContains(t, err, tc.errorContains)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRoomID, result)
			}
			r.set.AssertExpectations(t)
			r.roomRepo.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseRoomUnitSuite) TestReleaseRoom(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, roomID model.RoomID)
		expectError   bool
		errorType     error
		errorContains string
	}{
		{
			name: "Should release room successfully",
			setupMocks: func(r *resources, roomID model.RoomID) {
				r.roomRepo.On("ReleaseRoom", r.ctx, roomID).Return(nil).Once()
				r.set.On("Add", r.ctx, roomID).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Should return ErrReleaseRoom when repository fails",
			setupMocks: func(r *resources, roomID model.RoomID) {
				releaseError := errors.New("release error")
				r.roomRepo.On("ReleaseRoom", r.ctx, roomID).Return(releaseError).Once()
			},
			expectError:   true,
			errorType:     ErrReleaseRoom,
			errorContains: "release error",
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			roomID := validRoomID()
			tc.setupMocks(r, roomID)

			err := r.usecase.ReleaseRoom(r.ctx, roomID)

			if tc.expectError {
				assert.ErrorIs(t, err, tc.errorType)
				assert.ErrorContains(t, err, tc.errorContains)
			} else {
				assert.NoError(t, err)
			}
			r.roomRepo.AssertExpectations(t)
			if !tc.expectError {
				r.set.AssertExpectations(t)
			}
		})
	}
}

func TestUnitSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseRoomUnitSuite))
}
