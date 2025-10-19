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

type VoteRepository interface {
	GetRoomIDByCode(ctx context.Context, code string) (uuid.UUID, error)
	GetParticipantsEmbeddings(ctx context.Context, roomID uuid.UUID) ([]model.Embedding, error)
	GetSimilarMovies(ctx context.Context, queryEmbedding []float32, limit int) ([]*model.MovieMeta, error)
	Results(ctx context.Context, roomID uuid.UUID) ([]*model.Result, error)
	AddReactions(ctx context.Context, roomID uuid.UUID, reactions map[uuid.UUID]int) error
}

type Usecase struct {
	VoteRepository VoteRepository
}

func (u *Usecase) VotingBatch(ctx context.Context, n int, code string) ([]*model.MovieMeta, error) {
	roomID, err := u.VoteRepository.GetRoomIDByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, errors.Join(ErrInternal, err)
	}

	embeddings, err := u.VoteRepository.GetParticipantsEmbeddings(ctx, roomID)
	if err != nil {
		return nil, errors.Join(ErrInternal, err)
	}

	if len(embeddings) == 0 {
		return []*model.MovieMeta{}, ErrResourceNotFound
	}

	avgEmbedding := u.averageEmbeddings(embeddings)

	movies, err := u.VoteRepository.GetSimilarMovies(ctx, avgEmbedding, n)
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
	roomID, err := u.VoteRepository.GetRoomIDByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return nil, ErrResourceNotFound
		}
		return nil, errors.Join(ErrInternal, err)
	}

	results, err := u.VoteRepository.Results(ctx, roomID)
	if err != nil {
		return nil, errors.Join(ErrInternal, err)
	}

	return results, nil
}

func (u *Usecase) AddReaction(ctx context.Context, code string, reactions model.Reactions) error {
	roomID, err := u.VoteRepository.GetRoomIDByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return ErrResourceNotFound
		}
		return errors.Join(ErrInternal, err)
	}

	reactionMap := make(map[uuid.UUID]int)
	for movie, reaction := range reactions.Reactions {
		if reaction == model.PassReaction {
			reactionMap[movie.ID] = 1
		}
	}

	if len(reactionMap) == 0 {
		return nil
	}

	err = u.VoteRepository.AddReactions(ctx, roomID, reactionMap)
	if err != nil {
		return errors.Join(ErrInternal, err)
	}

	return nil
}
