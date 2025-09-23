package usecase_room

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
)

var (
	ErrCreate                 = errors.New("failed to create new room")
	ErrParticipate            = errors.New("failed to participate")
	ErrEnterRoom              = errors.New("failed to enter room")
	ErrBuildEmbedding         = errors.New("failed to create embedding")
	ErrFailedToStoreEmbedding = errors.New("failed to store embedding")
)

type Repository interface {
	CreateAndAquire(ctx context.Context, roomID model.RoomID) error
	FindAndAcquire(ctx context.Context) (model.RoomID, error)
	TryAcquire(ctx context.Context, roomID model.RoomID) error
	IsExistsRoomID(ctx context.Context, roomID model.RoomID) (bool, error)

	IsRoomAcquired(ctx context.Context, roomID model.RoomID) (bool, error)
	Participate(ctx context.Context, roomID model.RoomID, preference model.Preference) error
}

type IDCacheSet interface {
	Remove(ctx context.Context) (model.RoomID, error)
	Add(ctx context.Context, roomID model.RoomID) error
}

type Embedder interface {
	EmbedPreference(ctx context.Context, ID uuid.UUID, P model.Preference) error
}

type EmbeddingStorer interface {
	Store(ctx context.Context, roomID uuid.UUID, e model.Embedding) error
}

type Usecase struct {
	repo       Repository
	idCahceSet IDCacheSet

	embedder        Embedder
	embeddingStorer EmbeddingStorer
}

func New(repo Repository, embedder Embedder, embeddingStorer EmbeddingStorer, idCacheSet IDCacheSet) *Usecase {
	return &Usecase{
		repo:            repo,
		embedder:        embedder,
		embeddingStorer: embeddingStorer,
		idCahceSet:      idCacheSet,
	}
}

func (u *Usecase) AcquireRoom(ctx context.Context) (model.RoomID, error) {
	roomID, _ := u.idCahceSet.Remove(ctx)
	if roomID != model.EmptyRoomID {
		// Fast path
		if err := u.repo.TryAcquire(ctx, roomID); err == nil {
			return roomID, nil
		}
	}
	// Slower path
	roomID, _ = u.repo.FindAndAcquire(ctx)
	if roomID != model.EmptyRoomID {
		return roomID, nil
	}

	// Slowest path
	roomID, err := u.resolveRoomID(ctx)
	if err != nil {
		return model.EmptyRoomID, err
	}
	return roomID, u.repo.CreateAndAquire(ctx, roomID)
}

func (u *Usecase) ReleaseRoom(ctx context.Context, roomID model.RoomID) error {

	return nil
}

func (u *Usecase) Participate(ctx context.Context, roomID model.RoomID, p model.Preference) error {
	if err := u.repo.Participate(ctx, roomID, p); err != nil {
		return fmt.Errorf("%w:%w", ErrParticipate, err)
	}

	u.embedder.EmbedPreference(ctx, roomID.BuildUUID(), p)

	return nil
}

func (u *Usecase) IsRoomAcquired(ctx context.Context, roomID model.RoomID) (bool, error) {
	if ok, err := u.repo.IsRoomAcquired(ctx, roomID); !ok || err != nil {
		return false, err
	}
	return true, nil
}

func (u *Usecase) resolveRoomID(ctx context.Context) (model.RoomID, error) {
	var roomID model.RoomID
	for {
		roomID = u.buildRoomID()
		exists, err := u.repo.IsExistsRoomID(ctx, roomID)
		if err != nil {
			return model.EmptyRoomID, err
		}
		if !exists {
			break
		}
	}
	return roomID, nil
}

func (u *Usecase) buildRoomID() model.RoomID {
	const codeLen = 6
	var builder strings.Builder
	builder.Grow(codeLen)

	for range codeLen {
		builder.WriteByte(byte(rand.Intn(10)) + '0')
	}

	return model.RoomID(builder.String())
}
