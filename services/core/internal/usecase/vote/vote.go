package usecase_vote

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
)

var (
	ErrInternal         = errors.New("internal error")
	ErrResourceNotFound = errors.New("no such resource")
)

//go:generate mockery --name=VoteRepository --output=./mocks/vote/repository --filename=repository.go
type VoteRepository interface {
	RoomIDByCode(ctx context.Context, code string) (uuid.UUID, error)
	ParticipantsEmbeddings(ctx context.Context, roomID uuid.UUID) ([]model.Embedding, error)
	SimilarMovies(ctx context.Context, queryEmbedding []float32, limit int) ([]*model.MovieMeta, error)
	Results(ctx context.Context, roomID uuid.UUID) ([]*model.Result, error)
	AddReactions(ctx context.Context, roomID uuid.UUID, userID uuid.UUID, reactions map[uuid.UUID]int) error
	IsAllReady(ctx context.Context, roomID uuid.UUID) (bool, error)
}

//go:generate mockery --name=RoomUUIDer --output=./mocks/vote/roomuc --filename=roomuc.go
type RoomUUIDer interface {
	UUIDByCode(ctx context.Context, code string) (uuid.UUID, error)
}

type Usecase struct {
	VoteRepository VoteRepository
	RoomUUIDer     RoomUUIDer
}

func New(
	VoteRepository VoteRepository,
	RoomUUIDer RoomUUIDer,
) *Usecase {
	return &Usecase{
		VoteRepository: VoteRepository,
		RoomUUIDer:     RoomUUIDer,
	}
}

func (u *Usecase) VotingBatch(ctx context.Context, n int, code string) ([]*model.MovieMeta, error) {
	roomID, err := u.RoomUUIDer.UUIDByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	embeddings, err := u.VoteRepository.ParticipantsEmbeddings(ctx, roomID)
	if err != nil {
		return nil, errors.Join(ErrInternal, err)
	}

	if len(embeddings) == 0 {
		return []*model.MovieMeta{}, ErrResourceNotFound
	}

	avgEmbedding := u.averageEmbeddings(embeddings)

	movies, err := u.VoteRepository.SimilarMovies(ctx, avgEmbedding, n)
	if err != nil {
		return nil, errors.Join(ErrInternal, err)
	}

	return movies, nil
}

func (u *Usecase) averageEmbeddings(embeddings []model.Embedding) model.Embedding {
	if len(embeddings) == 0 {
		return nil
	}

	dim := len(embeddings[0])
	result := make([]float32, dim)

	for _, embedding := range embeddings {
		for i, value := range embedding {
			result[i] += value
		}
	}

	for i := range result {
		result[i] /= float32(len(embeddings))
	}

	return result
}

func (u *Usecase) Results(ctx context.Context, code string) ([]*model.Result, error) {
	roomID, err := u.RoomUUIDer.UUIDByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	results, err := u.VoteRepository.Results(ctx, roomID)
	if err != nil {
		return nil, errors.Join(ErrInternal, err)
	}

	return results, nil
}

// Passing userID to mark him as voted
// in order to make method idempotent
func (u *Usecase) AddReaction(ctx context.Context, code string, userID uuid.UUID, reactions model.Reactions) error {
	roomID, err := u.RoomUUIDer.UUIDByCode(ctx, code)
	if err != nil {
		return err
	}

	err = u.VoteRepository.AddReactions(ctx, roomID, userID, reactions.Reactions)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return ErrResourceNotFound
		}
		return errors.Join(ErrInternal, err)
	}

	return nil
}

func (u *Usecase) IsAllReady(ctx context.Context, code string) (bool, error) {
	roomID, err := u.RoomUUIDer.UUIDByCode(ctx, code)
	if err != nil {
		return false, err
	}

	ready, err := u.VoteRepository.IsAllReady(ctx, roomID)
	if err != nil {
		return false, errors.Join(ErrInternal, err)
	}
	return ready, err
}
