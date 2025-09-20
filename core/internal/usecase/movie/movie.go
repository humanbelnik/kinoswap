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
)

type Embedder interface {
	Embed(ctx context.Context, mm model.MovieMeta) ([]byte, error)
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
	Store(ctx context.Context, e model.Embedding) error
	Load(ctx context.Context, ID uuid.UUID) (model.Embedding, error)
	Delete(ctx context.Context, ID uuid.UUID) error
	KNN(ctx context.Context, K int, embedding model.Embedding) ([]uuid.UUID, error)
}

type RoomPreferenceService interface {
	GetPreferenceEmbedding(ctx context.Context, roomID model.RoomID) ([]byte, error)
}

type Usecase struct {
	embedder            Embedder
	metaRepository      MetaRepository
	embeddingRepository EmbeddingRepository
	roomPrefService     RoomPreferenceService
}

func New(
	embedder Embedder,
	meta MetaRepository,
	embeddingRepo EmbeddingRepository,
	roomPrefService RoomPreferenceService,
) *Usecase {
	return &Usecase{
		embedder:            embedder,
		metaRepository:      meta,
		embeddingRepository: embeddingRepo,
		roomPrefService:     roomPrefService,
	}
}

func (u *Usecase) TopK(ctx context.Context, roomID model.RoomID, K int) ([]*model.MovieMeta, error) {
	if K <= 0 {
		return nil, fmt.Errorf("%w: K must be positive", ErrInvalidInput)
	}

	// Get unite preference embedding (UPE) by roomID
	prefEmbedding, err := u.roomPrefService.GetPreferenceEmbedding(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToGetPreferenceEmbedding, err)
	}

	// Find KNN to UPE and return their IDs
	similarMovieIDs, err := u.embeddingRepository.KNN(ctx, K, model.Embedding{E: prefEmbedding})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrFailedToFindKNN, err)
	}

	// Fetch meta by IDs
	metaList, err := u.metaRepository.LoadByIDs(ctx, similarMovieIDs)
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

	// Create embedding model
	embedding := model.Embedding{
		ID: mm.ID,
		E:  embeddingBytes,
	}

	// Store meta in MetaRepository
	if err := u.metaRepository.Store(ctx, mm); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToStoreMeta, err)
	}

	// Store embedding in EmbeddingRepository
	if err := u.embeddingRepository.Store(ctx, embedding); err != nil {
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

	// Create updated embedding
	embedding := model.Embedding{
		ID: mm.ID,
		E:  embeddingBytes,
	}

	if err := u.embeddingRepository.Store(ctx, embedding); err != nil {
		return fmt.Errorf("failed to update embedding: %w", err)
	}

	return nil
}
