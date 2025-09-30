package usecase_room

import (
	"context"
	"errors"
	"testing"

	"github.com/humanbelnik/kinoswap/core/internal/model"
	cache_mocks "github.com/humanbelnik/kinoswap/core/mocks/cache"
	embedder_mocks "github.com/humanbelnik/kinoswap/core/mocks/embedder"
	repo_mocks "github.com/humanbelnik/kinoswap/core/mocks/repository"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type UsecaseRoomUnitSuite struct {
	suite.Suite

	usecase *Usecase

	roomRepo      *repo_mocks.RoomRepository
	embeddingRepo *repo_mocks.EmbeddingRepository
	set           *cache_mocks.EmptyRoomsIDSet
	embedder      *embedder_mocks.Embedder

	ctx context.Context
}

func validRoomID() model.RoomID {
	return model.RoomID("123456")
}

func (s *UsecaseRoomUnitSuite) BeforeEach(t provider.T) {
	s.roomRepo = repo_mocks.NewRoomRepository(t)
	s.embeddingRepo = repo_mocks.NewEmbeddingRepository(t)
	s.set = cache_mocks.NewEmptyRoomsIDSet(t)
	s.embedder = embedder_mocks.NewEmbedder(t)
	s.usecase = New(s.roomRepo, s.embedder, s.embeddingRepo, s.set)
	s.ctx = context.Background()
}

func (s *UsecaseRoomUnitSuite) TestIsRoomAcquired(t provider.T) {
	t.Run("Should return true when repository returns true", func(t provider.T) {
		roomID := validRoomID()

		s.roomRepo.On("IsRoomAcquired", s.ctx, roomID).Return(true, nil).Once()

		result, err := s.usecase.IsRoomAcquired(s.ctx, roomID)

		assert.NoError(t, err)
		assert.True(t, result)
		s.roomRepo.AssertExpectations(t)
	})

	t.Run("Should return false when repository returns false", func(t provider.T) {
		roomID := validRoomID()

		s.roomRepo.On("IsRoomAcquired", s.ctx, roomID).Return(false, nil).Once()

		result, err := s.usecase.IsRoomAcquired(s.ctx, roomID)

		assert.NoError(t, err)
		assert.False(t, result)
		s.roomRepo.AssertExpectations(t)
	})

	t.Run("Should return false and error when repository returns error", func(t provider.T) {
		roomID := validRoomID()
		expectedError := errors.New("repository error")

		s.roomRepo.On("IsRoomAcquired", s.ctx, roomID).Return(false, expectedError).Once()

		result, err := s.usecase.IsRoomAcquired(s.ctx, roomID)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.False(t, result)
		s.roomRepo.AssertExpectations(t)
	})
}

func (s *UsecaseRoomUnitSuite) TestParticipate(t provider.T) {
	t.Run("Should participate successfully", func(t provider.T) {
		roomID := validRoomID()
		preference := model.Preference{Text: "test preference"}

		s.roomRepo.On("AppendPreference", s.ctx, roomID, preference).Return(nil).Once()
		s.embedder.On("BuildPreferenceEmbedding", mock.Anything, preference).Return(model.Embedding{}, nil)
		s.embeddingRepo.On("Append", mock.Anything, roomID, mock.AnythingOfType("model.Embedding")).Return(nil)

		err := s.usecase.Participate(s.ctx, roomID, preference)

		assert.NoError(t, err)
		s.roomRepo.AssertExpectations(t)
	})

	t.Run("Should return ErrParticipate when repository fails", func(t provider.T) {
		roomID := validRoomID()
		preference := model.Preference{Text: "test preference"}
		repoError := errors.New("repository error")

		s.roomRepo.On("AppendPreference", s.ctx, roomID, preference).Return(repoError).Once()

		err := s.usecase.Participate(s.ctx, roomID, preference)

		assert.ErrorIs(t, err, ErrParticipate)
		assert.ErrorContains(t, err, repoError.Error())
		s.roomRepo.AssertExpectations(t)
	})
}
func (s *UsecaseRoomUnitSuite) TestAcquireRoom(t provider.T) {
	t.Run("Should acquire room successfully from cache", func(t provider.T) {
		roomID := validRoomID()

		s.set.On("Remove", s.ctx).Return(roomID, nil).Once()
		s.roomRepo.On("TryAcquire", s.ctx, roomID).Return(nil).Once()

		result, err := s.usecase.AcquireRoom(s.ctx)

		assert.NoError(t, err)
		assert.Equal(t, roomID, result)
		s.set.AssertExpectations(t)
		s.roomRepo.AssertExpectations(t)
	})

	t.Run("Should return ErrCreateRoom when all paths fail", func(t provider.T) {
		createError := errors.New("create error")

		s.set.On("Remove", s.ctx).Return(model.EmptyRoomID, nil).Once()
		s.roomRepo.On("FindAndAcquire", s.ctx).Return(model.EmptyRoomID, nil).Once()
		s.roomRepo.On("IsExistsRoomID", s.ctx, mock.AnythingOfType("model.RoomID")).Return(false, nil).Once()
		s.roomRepo.On("CreateAndAquire", s.ctx, mock.AnythingOfType("model.RoomID")).Return(createError).Once()

		_, err := s.usecase.AcquireRoom(s.ctx)

		assert.ErrorIs(t, err, ErrCreateRoom)
		assert.ErrorContains(t, err, createError.Error())
		s.set.AssertExpectations(t)
		s.roomRepo.AssertExpectations(t)
	})
}

func (s *UsecaseRoomUnitSuite) TestReleaseRoom(t provider.T) {
	t.Run("Should release room successfully", func(t provider.T) {
		roomID := validRoomID()

		s.roomRepo.On("ReleaseRoom", s.ctx, roomID).Return(nil).Once()
		s.set.On("Add", s.ctx, roomID).Return(nil).Once()

		err := s.usecase.ReleaseRoom(s.ctx, roomID)

		assert.NoError(t, err)
		s.roomRepo.AssertExpectations(t)
		s.set.AssertExpectations(t)
	})

	t.Run("Should return ErrReleaseRoom when repository fails", func(t provider.T) {
		roomID := validRoomID()
		releaseError := errors.New("release error")

		s.roomRepo.On("ReleaseRoom", s.ctx, roomID).Return(releaseError).Once()

		err := s.usecase.ReleaseRoom(s.ctx, roomID)

		assert.ErrorIs(t, err, ErrReleaseRoom)
		assert.ErrorContains(t, err, releaseError.Error())
		s.roomRepo.AssertExpectations(t)
	})
}

func TestUnitSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseRoomUnitSuite))
}
