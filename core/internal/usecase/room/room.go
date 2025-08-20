package usecase_room

import (
	"context"
	"errors"
	"fmt"

	"github.com/humanbelnik/kinoswap/core/internal/model"
)

var (
	ErrCreate        = errors.New("failed to create new room")
	ErrAddPreference = errors.New("failed to add preference")
)

type RoomStorage interface {
	AcquireRoom(ctx context.Context) (model.RoomID, error)
	AddPreference(p model.Preferece) error
}

type Usecase struct {
	storage RoomStorage
}

func New(storage RoomStorage) *Usecase {
	return &Usecase{
		storage: storage,
	}
}

func (uc *Usecase) AcquireRoom(ctx context.Context) (model.RoomID, error) {
	roomID, err := uc.storage.AcquireRoom(ctx)
	if err != nil {
		return model.EmptyRoomID, fmt.Errorf("%w:%w", ErrCreate, err)
	}

	return roomID, nil
}

func (uc *Usecase) AddPreference(p model.Preferece) error {
	if err := uc.storage.AddPreference(p); err != nil {
		return fmt.Errorf("%w:%w", ErrAddPreference, err)
	}

	return nil
}
