package usecase_vote

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	mocks "github.com/humanbelnik/kinoswap/core/mocks/repository"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
)

type UsecaseVoteUnitSuite struct {
	suite.Suite

	mockRepo *mocks.VoteRepository
	usecase  *Usecase
	ctx      context.Context
}

/*
'Object Mother' pattern example
aka cooks specific objects.
*/
func validRoomID() model.RoomID {
	return model.RoomID("123456")
}

func validMovieMeta() *model.MovieMeta {
	return &model.MovieMeta{
		ID:         uuid.New(),
		PosterLink: "link",
		Title:      "title",
		Genres:     []string{"genre1", "genre2"},
		Year:       2000,
		Rating:     5.5,
	}
}

func validMovieMetas(n int) []*model.MovieMeta {
	mms := make([]*model.MovieMeta, n)
	for i := range n {
		mms[i] = validMovieMeta()
	}
	return mms
}

func validVoteResult() model.VoteResult {
	mm := validMovieMeta()
	return model.VoteResult{
		Results: map[*model.MovieMeta]model.Reaction{
			mm: model.PassReaction,
		},
	}
}

func (s *UsecaseVoteUnitSuite) BeforeEach(t provider.T) {
	s.mockRepo = mocks.NewVoteRepository(t)
	s.usecase = New(s.mockRepo)
	s.ctx = context.Background()
}

func (s *UsecaseVoteUnitSuite) TestVote(t provider.T) {
	t.Run("Should save vote successfully", func(t provider.T) {
		roomID := validRoomID()
		voteResult := validVoteResult()

		s.mockRepo.On("AddVote", s.ctx, roomID, voteResult).
			Return(nil).Once()

		err := s.usecase.Vote(s.ctx, roomID, voteResult)

		assert.NoError(t, err)
		s.mockRepo.AssertExpectations(t)
	})

	t.Run("Should return error when repository fails", func(t provider.T) {
		roomID := validRoomID()
		voteResult := validVoteResult()
		repoError := ErrUnableToSaveVotes

		s.mockRepo.On("AddVote", s.ctx, roomID, voteResult).
			Return(repoError).Once()

		err := s.usecase.Vote(s.ctx, roomID, voteResult)

		assert.ErrorIs(t, err, ErrUnableToSaveVotes)
		s.mockRepo.AssertExpectations(t)
	})
}

func (s *UsecaseVoteUnitSuite) TestResults(t provider.T) {
	t.Run("Should return results successfully", func(t provider.T) {
		roomID := validRoomID()
		mmsExpected := validMovieMetas(5)

		s.mockRepo.On("LoadResults", s.ctx, roomID).Return(mmsExpected, nil).Once()

		mmsActual, err := s.usecase.Results(s.ctx, roomID)

		assert.NoError(t, err)
		s.mockRepo.AssertExpectations(t)
		assert.ElementsMatch(t, mmsExpected, mmsActual)
	})

	t.Run("Should return error when repository fails", func(t provider.T) {
		roomID := validRoomID()
		repoError := ErrUnableToGetResults

		s.mockRepo.On("LoadResults", s.ctx, roomID).Return(nil, repoError).Once()

		mms, err := s.usecase.Results(s.ctx, roomID)

		assert.ErrorIs(t, err, ErrUnableToGetResults)
		assert.Nil(t, mms)
		s.mockRepo.AssertExpectations(t)
	})
}

func TestUnitSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseVoteUnitSuite))
}
