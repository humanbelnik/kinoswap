package usecase_vote

import (
	"context"
	"errors"
	"fmt"

	"github.com/humanbelnik/kinoswap/core/internal/model"
)

var (
	ErrUnableToSaveVotes  = errors.New("unable to vote")
	ErrUnableToGetResults = errors.New("unable to get results")
)

type VoteRepository interface {
	AddVote(ctx context.Context, roomID model.RoomID, results model.VoteResult) error
	LoadResults(ctx context.Context, roomID model.RoomID) ([]*model.MovieMeta, error)
}

type Usecase struct {
	voteRepository VoteRepository
}

func New(
	r VoteRepository,
) *Usecase {
	return &Usecase{
		voteRepository: r,
	}
}

func (u *Usecase) Vote(ctx context.Context, roomID model.RoomID, results model.VoteResult) error {
	if err := u.voteRepository.AddVote(ctx, roomID, results); err != nil {
		return fmt.Errorf("%w : %w", ErrUnableToSaveVotes, err)
	}

	return nil
}

func (u *Usecase) Results(ctx context.Context, roomID model.RoomID) ([]*model.MovieMeta, error) {
	results, err := u.voteRepository.LoadResults(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrUnableToGetResults, err)
	}

	return results, nil
}
