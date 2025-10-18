//go:build unit

package usecase_vote

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	mocks "github.com/humanbelnik/kinoswap/core/mocks/vote/repository"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
)

type UsecaseVoteUnitSuite struct {
	suite.Suite
}

type resources struct {
	mockRepo *mocks.VoteRepository
	usecase  *Usecase
	ctx      context.Context
}

func initResources(t provider.T) *resources {
	repo := mocks.NewVoteRepository(t)
	return &resources{
		mockRepo: repo,
		usecase:  New(repo),
		ctx:      context.Background(),
	}
}

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

func (suite *UsecaseVoteUnitSuite) TestVote(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		setupMocks  func(r *resources, roomID model.RoomID, voteResult model.VoteResult)
		expectError bool
	}{
		{
			name: "Should save vote successfully",
			setupMocks: func(r *resources, roomID model.RoomID, voteResult model.VoteResult) {
				r.mockRepo.On("AddVote", r.ctx, roomID, voteResult).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, roomID model.RoomID, voteResult model.VoteResult) {
				r.mockRepo.On("AddVote", r.ctx, roomID, voteResult).Return(ErrUnableToSaveVotes).Once()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			roomID := validRoomID()
			voteResult := validVoteResult()
			tc.setupMocks(r, roomID, voteResult)

			err := r.usecase.Vote(r.ctx, roomID, voteResult)

			if tc.expectError {
				assert.ErrorIs(t, err, ErrUnableToSaveVotes)
			} else {
				assert.NoError(t, err)
			}
			r.mockRepo.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseVoteUnitSuite) TestResults(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, roomID model.RoomID, expectedMetas []*model.MovieMeta)
		expectError   bool
		expectedMetas []*model.MovieMeta
	}{
		{
			name: "Should return results successfully",
			setupMocks: func(r *resources, roomID model.RoomID, expectedMetas []*model.MovieMeta) {
				r.mockRepo.On("LoadResults", r.ctx, roomID).Return(expectedMetas, nil).Once()
			},
			expectError:   false,
			expectedMetas: validMovieMetas(5),
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, roomID model.RoomID, expectedMetas []*model.MovieMeta) {
				r.mockRepo.On("LoadResults", r.ctx, roomID).Return(nil, ErrUnableToGetResults).Once()
			},
			expectError:   true,
			expectedMetas: nil,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			roomID := validRoomID()
			tc.setupMocks(r, roomID, tc.expectedMetas)

			mmsActual, err := r.usecase.Results(r.ctx, roomID)

			if tc.expectError {
				assert.ErrorIs(t, err, ErrUnableToGetResults)
				assert.Nil(t, mmsActual)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tc.expectedMetas, mmsActual)
			}
			r.mockRepo.AssertExpectations(t)
		})
	}
}

func TestUnitSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseVoteUnitSuite))
}
