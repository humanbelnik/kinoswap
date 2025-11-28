//go:build !integration
// +build !integration

package usecase_vote

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	mocks_repo "github.com/humanbelnik/kinoswap/core/internal/usecase/vote/mocks/vote/repository"
	mocks_room "github.com/humanbelnik/kinoswap/core/internal/usecase/vote/mocks/vote/roomuc"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
)

type UsecaseVoteUnitSuite struct {
	suite.Suite
}

type resources struct {
	mockRepo   *mocks_repo.VoteRepository
	mockRoomUC *mocks_room.RoomUUIDer
	usecase    *Usecase
	ctx        context.Context
}

func initResources(t provider.T) *resources {
	repo := mocks_repo.NewVoteRepository(t)
	roomUC := mocks_room.NewRoomUUIDer(t)
	return &resources{
		mockRepo:   repo,
		mockRoomUC: roomUC,
		usecase:    New(repo, roomUC),
		ctx:        context.Background(),
	}
}

func validRoomID() uuid.UUID {
	return uuid.New()
}

func validCode() string {
	return "ABC123"
}

func validUserID() uuid.UUID {
	return uuid.New()
}

func validEmbeddings(n int) []model.Embedding {
	embeddings := make([]model.Embedding, n)
	for i := range n {
		embeddings[i] = model.Embedding{0.1, 0.2, 0.3}
	}
	return embeddings
}

func validMovieMeta() model.MovieMeta {
	return model.MovieMeta{
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
		meta := validMovieMeta()
		mms[i] = &meta
	}
	return mms
}

func validReactions() model.Reactions {
	return model.Reactions{
		Reactions: map[uuid.UUID]model.Reaction{
			uuid.New(): model.PassReaction,
			uuid.New(): model.SmashReaction,
		},
	}
}

func validResults() []*model.Result {
	return []*model.Result{
		{
			MM:    validMovieMeta(),
			Likes: 5,
		},
		{
			MM:    validMovieMeta(),
			Likes: 3,
		},
	}
}

func (suite *UsecaseVoteUnitSuite) TestVotingBatch(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		setupMocks     func(r *resources, code string, roomID uuid.UUID)
		expectError    bool
		expectedMovies []*model.MovieMeta
	}{
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string, roomID uuid.UUID) {
				r.mockRoomUC.On("UUIDByCode", r.ctx, code).Return(roomID, nil).Once()
				r.mockRepo.On("ParticipantsEmbeddings", r.ctx, roomID).Return(nil, ErrInternal).Once()
			},
			expectError:    true,
			expectedMovies: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validCode()
			roomID := validRoomID()
			tc.setupMocks(r, code, roomID)

			movies, err := r.usecase.VotingBatch(r.ctx, 10, code)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedMovies, movies)
			}
			r.mockRepo.AssertExpectations(t)
			r.mockRoomUC.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseVoteUnitSuite) TestResults(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		setupMocks      func(r *resources, code string, roomID uuid.UUID)
		expectError     bool
		expectedResults []*model.Result
	}{
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string, roomID uuid.UUID) {
				r.mockRoomUC.On("UUIDByCode", r.ctx, code).Return(roomID, nil).Once()
				r.mockRepo.On("Results", r.ctx, roomID).Return(nil, ErrInternal).Once()
			},
			expectError:     true,
			expectedResults: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validCode()
			roomID := validRoomID()
			tc.setupMocks(r, code, roomID)

			results, err := r.usecase.Results(r.ctx, code)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResults, results)
			}
			r.mockRepo.AssertExpectations(t)
			r.mockRoomUC.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseVoteUnitSuite) TestAddReaction(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		setupMocks  func(r *resources, code string, roomID uuid.UUID, userID uuid.UUID, reactions model.Reactions)
		expectError bool
	}{
		{
			name: "Should add reactions successfully",
			setupMocks: func(r *resources, code string, roomID uuid.UUID, userID uuid.UUID, reactions model.Reactions) {
				r.mockRoomUC.On("UUIDByCode", r.ctx, code).Return(roomID, nil).Once()
				r.mockRepo.On("AddReactions", r.ctx, roomID, userID, reactions.Reactions).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string, roomID uuid.UUID, userID uuid.UUID, reactions model.Reactions) {
				r.mockRoomUC.On("UUIDByCode", r.ctx, code).Return(roomID, nil).Once()
				r.mockRepo.On("AddReactions", r.ctx, roomID, userID, reactions.Reactions).Return(ErrInternal).Once()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validCode()
			roomID := validRoomID()
			userID := validUserID()
			reactions := validReactions()
			tc.setupMocks(r, code, roomID, userID, reactions)

			err := r.usecase.AddReaction(r.ctx, code, userID, reactions)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			r.mockRepo.AssertExpectations(t)
			r.mockRoomUC.AssertExpectations(t)
		})
	}
}

func (suite *UsecaseVoteUnitSuite) TestIsAllReady(t provider.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		setupMocks    func(r *resources, code string, roomID uuid.UUID)
		expectError   bool
		expectedReady bool
	}{
		{
			name: "Should return ready status successfully",
			setupMocks: func(r *resources, code string, roomID uuid.UUID) {
				r.mockRoomUC.On("UUIDByCode", r.ctx, code).Return(roomID, nil).Once()
				r.mockRepo.On("IsAllReady", r.ctx, roomID).Return(true, nil).Once()
			},
			expectError:   false,
			expectedReady: true,
		},
		{
			name: "Should return error when repository fails",
			setupMocks: func(r *resources, code string, roomID uuid.UUID) {
				r.mockRoomUC.On("UUIDByCode", r.ctx, code).Return(roomID, nil).Once()
				r.mockRepo.On("IsAllReady", r.ctx, roomID).Return(false, ErrInternal).Once()
			},
			expectError:   true,
			expectedReady: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t provider.T) {
			t.Parallel()
			r := initResources(t)
			code := validCode()
			roomID := validRoomID()
			tc.setupMocks(r, code, roomID)

			ready, err := r.usecase.IsAllReady(r.ctx, code)

			if tc.expectError {
				assert.Error(t, err)
				assert.False(t, ready)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedReady, ready)
			}
			r.mockRepo.AssertExpectations(t)
			r.mockRoomUC.AssertExpectations(t)
		})
	}
}

func TestUnitSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseVoteUnitSuite))
}
