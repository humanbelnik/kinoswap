package usecase_room

import (
	"context"
	"errors"
	"fmt"

	"github.com/humanbelnik/kinoswap/core/internal/model"
)

var (
	ErrCreate                 = errors.New("failed to create new room")
	ErrParticipate            = errors.New("failed to participate")
	ErrEnterRoom              = errors.New("failed to enter room")
	ErrBuildEmbedding         = errors.New("failed to create embedding")
	ErrFailedToStoreEmbedding = errors.New("failed to store embedding")
)

type RoomStorage interface {
	AcquireRoom(ctx context.Context) (model.RoomID, error)
	Participate(ctx context.Context, roomID model.RoomID, p model.Preference) error
	IsRoomAcquired(ctx context.Context, roomID model.RoomID) (bool, error)
}

type Embedder interface {
	Embed(ctx context.Context, v any) ([]byte, error)
}

type EmbeddingStorer interface {
	Store(ctx context.Context, roomID model.RoomID, e model.Embedding) error
}

type Usecase struct {
	storage         RoomStorage
	embedder        Embedder
	embeddingStorer EmbeddingStorer
}

func New(storage RoomStorage, embedder Embedder, embeddingStorer EmbeddingStorer) *Usecase {
	return &Usecase{
		storage:         storage,
		embedder:        embedder,
		embeddingStorer: embeddingStorer,
	}
}

func (u *Usecase) AcquireRoom(ctx context.Context) (model.RoomID, error) {
	roomID, err := u.storage.AcquireRoom(ctx)
	if err != nil {
		return model.EmptyRoomID, fmt.Errorf("%w:%w", ErrCreate, err)
	}

	return roomID, nil
}

func (u *Usecase) ReleaseRoom(ctx context.Context, roomID model.RoomID) error {

	return nil
}

func (u *Usecase) Participate(ctx context.Context, roomID model.RoomID, p model.Preference) error {
	if err := u.storage.Participate(ctx, roomID, p); err != nil {
		return fmt.Errorf("%w:%w", ErrParticipate, err)
	}

	prefEmbedding, err := u.embedder.Embed(ctx, p)
	if err != nil {
		return fmt.Errorf("%w:%w", ErrBuildEmbedding, err)
	}

	if err := u.embeddingStorer.Store(ctx, roomID, model.Embedding{E: prefEmbedding}); err != nil {
		return fmt.Errorf("%w:%w", ErrFailedToStoreEmbedding, err)
	}

	return nil
}

func (u *Usecase) IsRoomAcquired(ctx context.Context, roomID model.RoomID) (bool, error) {
	return u.storage.IsRoomAcquired(ctx, roomID)
}
