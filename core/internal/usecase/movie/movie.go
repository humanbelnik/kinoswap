package usecase_movie

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
)

var (
	ErrFailedToFetchCards             = errors.New("failed to fetch cards")
	ErrFailedToGetEmbedding           = errors.New("failed to get embedding")
	ErrFailedToStoreMeta              = errors.New("failed to store meta")
	ErrFailedToStoreEmbedding         = errors.New("failed to store embedding")
	ErrFailedToLoadMeta               = errors.New("failed to load meta")
	ErrFailedToFindKNN                = errors.New("failed to find KNN")
	ErrInvalidInput                   = errors.New("invalid input")
	ErrRoomNotFound                   = errors.New("room not found")
	ErrFailedToGetPreferenceEmbedding = errors.New("failed to get preference embedding")
	ErrTypeCastFailed                 = errors.New("id type cast failed")
)

type Embedder interface {
	Embed(ctx context.Context, v any) ([]byte, error)
}

type MetaRepository interface {
	Store(ctx context.Context, mm model.MovieMeta) error
	Load(ctx context.Context) ([]*model.MovieMeta, error)
	LoadByID(ctx context.Context, ID uuid.UUID) (model.MovieMeta, error)
	LoadByIDs(ctx context.Context, IDs []uuid.UUID) ([]*model.MovieMeta, error)
	Update(ctx context.Context, mm model.MovieMeta) error
	DeleteByID(ctx context.Context, ID uuid.UUID) error
}

type EmbeddingRepository interface {
	Store(ctx context.Context, ID model.EID, E model.Embedding) error
	Load(ctx context.Context, ID model.EID) (model.Embedding, error)
	Delete(ctx context.Context, ID model.EID) error
	KNN(ctx context.Context, K int, E model.Embedding) ([]model.EID, error)
}

type Usecase struct {
	embedder            Embedder
	metaRepository      MetaRepository
	embeddingRepository EmbeddingRepository
}

func New(
	embedder Embedder,
	meta MetaRepository,
	embeddingRepo EmbeddingRepository,

) *Usecase {
	return &Usecase{
		embedder:            embedder,
		metaRepository:      meta,
		embeddingRepository: embeddingRepo,
	}
}

func (u *Usecase) TopK(ctx context.Context, roomID model.RoomID, K int) ([]*model.MovieMeta, error) {
	if K <= 0 {
		return nil, fmt.Errorf("%w: K must be positive", ErrInvalidInput)
	}

	// Get unite preference embedding (UPE) by roomID
	prefEmbedding, err := u.embeddingRepository.Load(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGetPreferenceEmbedding, err)
	}

	// Find KNN to UPE and return their IDs
	similarMovieIDs, err := u.embeddingRepository.KNN(ctx, K, prefEmbedding)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToFindKNN, err)
	}

	uuids := make([]uuid.UUID, len(similarMovieIDs))

	for i, id := range similarMovieIDs {
		if v, ok := id.(uuid.UUID); ok {
			uuids[i] = v
		} else {
			return nil, ErrTypeCastFailed
		}
	}

	// Fetch meta by IDs
	metaList, err := u.metaRepository.LoadByIDs(ctx, uuids)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToLoadMeta, err)
	}

	return metaList, nil
}

func (u *Usecase) Upload(ctx context.Context, mm model.MovieMeta) error {
	if mm.ID == uuid.Nil {
		return fmt.Errorf("%w: movie ID cannot be nil", ErrInvalidInput)
	}

	if mm.Title == "" {
		return fmt.Errorf("%w: movie title cannot be empty", ErrInvalidInput)
	}

	// Get its embedding
	embeddingBytes, err := u.embedder.Embed(ctx, mm)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToGetEmbedding, err)
	}

	// Store meta in MetaRepository
	if err := u.metaRepository.Store(ctx, mm); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToStoreMeta, err)
	}

	// Store embedding in EmbeddingRepository
	if err := u.embeddingRepository.Store(ctx, mm.ID, model.Embedding(embeddingBytes)); err != nil {
		// Attempt to cleanup meta if embedding storage fails
		if deleteErr := u.metaRepository.DeleteByID(ctx, mm.ID); deleteErr != nil {
			return fmt.Errorf(
				"failed to store embedding and cleanup meta: %w, cleanup error: %w",
				err, deleteErr,
			)
		}
		return fmt.Errorf("%w: %w", ErrFailedToStoreEmbedding, err)
	}

	return nil
}

func (u *Usecase) GetMovieByID(ctx context.Context, id uuid.UUID) (model.MovieMeta, error) {
	if id == uuid.Nil {
		return model.MovieMeta{}, fmt.Errorf("%w: movie ID cannot be nil", ErrInvalidInput)
	}

	meta, err := u.metaRepository.LoadByID(ctx, id)
	if err != nil {
		return model.MovieMeta{}, fmt.Errorf("%w: %w", ErrFailedToLoadMeta, err)
	}

	return meta, nil
}

func (u *Usecase) DeleteMovie(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("%w: movie ID cannot be nil", ErrInvalidInput)
	}

	// Delete embedding first
	if err := u.embeddingRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete embedding: %w", err)
	}

	// Delete meta
	if err := u.metaRepository.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete meta: %w", err)
	}

	return nil
}

func (u *Usecase) UpdateMovie(ctx context.Context, mm model.MovieMeta) error {
	if mm.ID == uuid.Nil {
		return fmt.Errorf("%w: movie ID cannot be nil", ErrInvalidInput)
	}

	// Update meta first
	if err := u.metaRepository.Update(ctx, mm); err != nil {
		return fmt.Errorf("failed to update meta: %w", err)
	}

	// Regenerate and update embedding
	embeddingBytes, err := u.embedder.Embed(ctx, mm)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToGetEmbedding, err)
	}

	if err := u.embeddingRepository.Store(ctx, mm.ID, model.Embedding(embeddingBytes)); err != nil {
		return fmt.Errorf("failed to update embedding: %w", err)
	}

	return nil
}

func (u *Usecase) Load(ctx context.Context) ([]*model.MovieMeta, error) {
	mm, err := u.metaRepository.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrFailedToLoadMeta, err)
	}
	return mm, nil
}
