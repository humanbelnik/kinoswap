package usecase_room

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/humanbelnik/kinoswap/core/internal/model"
)

var (
	ErrCreateRoom             = errors.New("failed to create new room")
	ErrParticipate            = errors.New("failed to participate")
	ErrEnterRoom              = errors.New("failed to enter room")
	ErrBuildEmbedding         = errors.New("failed to create embedding")
	ErrFailedToStoreEmbedding = errors.New("failed to store embedding")
	ErrReleaseRoom            = errors.New("failed to release room")
	ErrAddToCache             = errors.New("failed to add roomID to a cache")
)

type RoomRepository interface {
	CreateAndAquire(ctx context.Context, roomID model.RoomID) error
	FindAndAcquire(ctx context.Context) (model.RoomID, error)
	TryAcquire(ctx context.Context, roomID model.RoomID) error
	IsExistsRoomID(ctx context.Context, roomID model.RoomID) (bool, error)
	IsRoomAcquired(ctx context.Context, roomID model.RoomID) (bool, error)
	AppendPreference(ctx context.Context, roomID model.RoomID, preference model.Preference) error
	ReleaseRoom(ctx context.Context, roomID model.RoomID) error
}

type EmptyRoomsIDSet interface {
	Remove(ctx context.Context) (model.RoomID, error)
	Add(ctx context.Context, roomID model.RoomID) error
}

type Embedder interface {
	BuildPreferenceEmbedding(ctx context.Context, p model.Preference) (model.Embedding, error)
}

type EmbeddingRepository interface {
	Append(ctx context.Context, roomID model.RoomID, e model.Embedding) error
}

type Usecase struct {
	repo       RoomRepository
	idCahceSet EmptyRoomsIDSet

	embedder         Embedder
	embeddingStorage EmbeddingRepository
}

func New(repo RoomRepository, embedder Embedder, embeddingStorage EmbeddingRepository, idCacheSet EmptyRoomsIDSet) *Usecase {
	return &Usecase{
		repo:             repo,
		embedder:         embedder,
		embeddingStorage: embeddingStorage,
		idCahceSet:       idCacheSet,
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

	err = u.repo.CreateAndAquire(ctx, roomID)
	if err != nil {
		err = fmt.Errorf("%w : %w", ErrCreateRoom, err)
	}
	return roomID, err
}

func (u *Usecase) ReleaseRoom(ctx context.Context, roomID model.RoomID) error {
	if err := u.repo.ReleaseRoom(ctx, roomID); err != nil {
		return fmt.Errorf("%w : %w", ErrReleaseRoom, err)
	}

	// Not critical
	if err := u.idCahceSet.Add(ctx, roomID); err != nil {
		return fmt.Errorf("%w : %w", ErrAddToCache, err)
	}
	return nil
}

func (u *Usecase) Participate(ctx context.Context, roomID model.RoomID, p model.Preference) error {
	if err := u.repo.AppendPreference(ctx, roomID, p); err != nil {
		return fmt.Errorf("%w:%w", ErrParticipate, err)
	}

	go func() {
		/*
			Don't use parent HTTP context on async tasks.
			Parent context cancels when response is made.
		*/
		ctx := context.Background()
		e, err := u.embedder.BuildPreferenceEmbedding(ctx, p)
		if err != nil {
			//! Logging here
			return
		}
		if err := u.embeddingStorage.Append(ctx, roomID, e); err != nil {
			//! Logging here
			return
		}
	}()

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
