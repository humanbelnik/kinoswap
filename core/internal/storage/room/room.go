package storage_room

import (
	"context"
	"math/rand"
	"strings"

	"github.com/humanbelnik/kinoswap/core/internal/model"
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

type Storage struct {
	repo       Repository
	idCahceSet IDCacheSet
}

func New(
	repo Repository,
	idCahceSet IDCacheSet,
) *Storage {
	return &Storage{
		repo:       repo,
		idCahceSet: idCahceSet,
	}
}

// At first try a fast pass:
// Get ID from cached set
//
// If there're none of them or it became acuired, go slow:
// acuire room with status 'free' in storage.
//
// At last, if all existing rooms are already acuired, create new one.
func (s *Storage) AcquireRoom(ctx context.Context) (model.RoomID, error) {
	roomID, _ := s.idCahceSet.Remove(ctx)
	if roomID != model.EmptyRoomID {
		// Fast path
		if err := s.repo.TryAcquire(ctx, roomID); err == nil {
			return roomID, nil
		}
	}
	// Slower path
	roomID, _ = s.repo.FindAndAcquire(ctx)
	if roomID != model.EmptyRoomID {
		return roomID, nil
	}

	// Slowest path
	roomID, err := s.resolveRoomID(ctx)
	if err != nil {
		return model.EmptyRoomID, err
	}
	return roomID, s.repo.CreateAndAquire(ctx, roomID)
}
func (s *Storage) resolveRoomID(ctx context.Context) (model.RoomID, error) {
	var roomID model.RoomID
	for {
		roomID = s.buildRoomID()
		exists, err := s.repo.IsExistsRoomID(ctx, roomID)
		if err != nil {
			return model.EmptyRoomID, err
		}
		if !exists {
			break
		}
	}
	return roomID, nil
}

func (s *Storage) buildRoomID() model.RoomID {
	const codeLen = 6
	var builder strings.Builder
	builder.Grow(codeLen)

	for range codeLen {
		builder.WriteByte(byte(rand.Intn(10)) + '0')
	}

	return model.RoomID(builder.String())
}

func (s *Storage) Participate(ctx context.Context, roomID model.RoomID, p model.Preference) error {
	return s.repo.Participate(ctx, roomID, p)
}
func (s *Storage) IsRoomAcquired(ctx context.Context, roomID model.RoomID) (bool, error) {
	if ok, err := s.repo.IsRoomAcquired(ctx, roomID); !ok || err != nil {
		return false, err
	}
	return true, nil
}
