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
	EmbedMovie(ctx context.Context, ID uuid.UUID, v model.MovieMeta) error
}

type Repository interface {
	Store(ctx context.Context, mm model.MovieMeta) error
	Load(ctx context.Context) ([]*model.MovieMeta, error)
	LoadByID(ctx context.Context, ID uuid.UUID) (model.MovieMeta, error)
	LoadByIDs(ctx context.Context, IDs []uuid.UUID) ([]*model.MovieMeta, error)
	Update(ctx context.Context, mm model.MovieMeta) error
	DeleteByID(ctx context.Context, ID uuid.UUID) error
	UpdateEmbedding(ctx context.Context, ID uuid.UUID, e model.Embedding) error

	KNN(ctx context.Context, k int, e model.Embedding) ([]*model.MovieMeta, error)
}

type EmbeddingRepository interface {
	Load(ctx context.Context, ID uuid.UUID) (model.Embedding, error)
}

type Usecase struct {
	embedder            Embedder
	repository          Repository
	embeddingRepository EmbeddingRepository
}

func New(
	embedder Embedder,
	meta Repository,
	embeddingRepo EmbeddingRepository,

) *Usecase {
	return &Usecase{
		embedder:            embedder,
		repository:          meta,
		embeddingRepository: embeddingRepo,
	}
}

func (u *Usecase) TopK(ctx context.Context, roomID model.RoomID, K int) ([]*model.MovieMeta, error) {
	if K <= 0 {
		return nil, fmt.Errorf("%w: K must be positive", ErrInvalidInput)
	}

	// Get unite preference embedding (UPE) by roomID
	prefEmbedding, err := u.embeddingRepository.Load(ctx, roomID.BuildUUID())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGetPreferenceEmbedding, err)
	}

	// Find KNN to UPE and return their IDs
	mm, err := u.repository.KNN(ctx, K, prefEmbedding)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToFindKNN, err)
	}

	return mm, nil
}

func (u *Usecase) Upload(ctx context.Context, mm model.MovieMeta) error {
	if mm.Title == "" {
		return fmt.Errorf("%w: movie title cannot be empty", ErrInvalidInput)
	}

	// Store meta in MetaRepository
	if err := u.repository.Store(ctx, mm); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToStoreMeta, err)
	}

	//TODO
	// Call async
	u.embedder.EmbedMovie(ctx, mm.ID, mm)

	return nil
}

func (u *Usecase) GetMovieByID(ctx context.Context, id uuid.UUID) (model.MovieMeta, error) {
	meta, err := u.repository.LoadByID(ctx, id)
	if err != nil {
		return model.MovieMeta{}, fmt.Errorf("%w: %w", ErrFailedToLoadMeta, err)
	}

	return meta, nil
}

func (u *Usecase) DeleteMovie(ctx context.Context, id uuid.UUID) error {
	if err := u.repository.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("failed to delete meta: %w", err)
	}

	return nil
}

func (u *Usecase) UpdateMovie(ctx context.Context, mm model.MovieMeta) error {
	if err := u.repository.Update(ctx, mm); err != nil {
		return fmt.Errorf("failed to update meta: %w", err)
	}

	//TODO
	// Call async
	u.embedder.EmbedMovie(ctx, mm.ID, mm)
	return nil
}

func (u *Usecase) Load(ctx context.Context) ([]*model.MovieMeta, error) {
	mm, err := u.repository.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrFailedToLoadMeta, err)
	}
	return mm, nil
}

func (u *Usecase) AddEmbedding(ctx context.Context, ID uuid.UUID, e model.Embedding) error {
	if err := u.repository.UpdateEmbedding(ctx, ID, e); err != nil {
		return fmt.Errorf("%w:%w", ErrFailedToStoreEmbedding, err)
	}

	return nil
}
