package usecase_room

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
)

var (
	ErrCodeConflict     = errors.New("code conflict")
	ErrRoomsUnavailable = errors.New("no available rooms")
	ErrInternal         = errors.New("internal error")
	ErrResourceNotFound = errors.New("no such resource")
)

//go:generate mockery --name=RoomRepository --output=./mocks/room/repository --filename=repository.go
type RoomRepository interface {
	CreateAndBook(ctx context.Context, room model.Room, ownerID uuid.UUID) error
	IsOwner(ctx context.Context, code string, ownerID uuid.UUID) (bool, error)
	DeleteByCode(ctx context.Context, code string) error
	StatusByCode(ctx context.Context, code string) (string, error)
	SetStatusByCode(ctx context.Context, code string, status string) error
	AddPreferenceEmbedding(ctx context.Context, code string, userID uuid.UUID, prefEmbedding model.Embedding) error
	ParticipantsCount(ctx context.Context, code string) (int, error)
	IsParticipant(ctx context.Context, code string, userID uuid.UUID) (bool, error)
	UUIDByCode(ctx context.Context, code string) (uuid.UUID, error)

	CleanupOrphantRooms(ctx context.Context, lobbiesDeadline, votingDeadline time.Duration) error
}

//go:generate mockery --name=Embedder --output=./mocks/room/Embedder --filename=Embedder.go
type Embedder interface {
	BuildPreferenceEmbedding(ctx context.Context, p model.Preference) (model.Embedding, error)
}

type Usecase struct {
	RoomRepository RoomRepository
	Embedder       Embedder

	// Used to make periodic stuff on every Nth cleanup
	cleanupPeriod int
	booksCount    int
}

func New(
	RoomRepository RoomRepository,
	Embedder Embedder,
	cleanup int,
) *Usecase {
	if cleanup <= 0 {
		cleanup = 20 /* default */
	}

	return &Usecase{
		RoomRepository: RoomRepository,
		Embedder:       Embedder,
		cleanupPeriod:  cleanup,
	}
}

// Owner token must be set on a client in order to be able to do 'owner ops'
func (u *Usecase) Book(ctx context.Context) (roomCode string, ownerToken string, err error) {
	ownerID := u.resolveOwnerToken()

	// Cleanup orphant rooms
	u.booksCount++
	if u.booksCount%u.cleanupPeriod == 0 {
		if err := u.RoomRepository.CleanupOrphantRooms(ctx, time.Minute*5 /* Lobbies */, time.Minute*10 /* Votings */); err != nil {
			return "", "", errors.Join(ErrInternal, err)
		}
	}

	roomCode, err = u.createRoomLobby(ctx, ownerID)
	if err != nil {
		return "", "", err
	}
	return roomCode, ownerID.String(), nil
}

// Assuming that codes can conflict.
// Retrying...
func (u *Usecase) createRoomLobby(ctx context.Context, ownerID uuid.UUID) (string, error) {
	var retries = 3
	for retries > 0 {
		code := u.buildRoomCode()
		if err := u.RoomRepository.CreateAndBook(ctx, model.Room{
			ID:         uuid.New(),
			PublicCode: code,
			Status:     model.StatusLobby,
		}, ownerID); err != nil {
			if errors.Is(err, ErrCodeConflict) {
				retries--
			} else {
				return "", errors.Join(ErrInternal, err)
			}
		} else {
			return code, nil
		}
	}
	return "", ErrRoomsUnavailable
}

func (u *Usecase) resolveOwnerToken() uuid.UUID {
	return uuid.New()
}

func (u *Usecase) buildRoomCode() string {
	const codeLen = 6
	var builder strings.Builder
	builder.Grow(codeLen)

	for range codeLen {
		builder.WriteByte(byte(rand.Intn(10)) + '0')
	}

	return builder.String()
}

func (u *Usecase) IsOwner(ctx context.Context, code string, ownerID string) (bool, error) {
	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return false, errors.Join(ErrInternal, err)
	}

	isOwner, err := u.RoomRepository.IsOwner(ctx, code, ownerUUID)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return false, ErrResourceNotFound
		}
		return false, errors.Join(ErrInternal, err)
	}

	return isOwner, nil
}

func (u *Usecase) IsParticipant(ctx context.Context, code string, userID string) (bool, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return false, errors.Join(ErrInternal, err)
	}

	isParticipant, err := u.RoomRepository.IsParticipant(ctx, code, userUUID)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return false, ErrResourceNotFound
		}
		return false, errors.Join(ErrInternal, err)
	}

	return isParticipant, nil
}

func (u *Usecase) Free(ctx context.Context, code string) error {
	if err := u.RoomRepository.DeleteByCode(ctx, code); err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return ErrResourceNotFound
		}
		return errors.Join(ErrInternal, err)
	}
	return nil
}

func (u *Usecase) Status(ctx context.Context, code string) (string, error) {
	status, err := u.RoomRepository.StatusByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return "", ErrResourceNotFound
		}
		return "", errors.Join(ErrInternal, err)
	}
	return status, nil
}

func (u *Usecase) SetStatus(ctx context.Context, code string, status string) error {
	err := u.RoomRepository.SetStatusByCode(ctx, code, status)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return ErrResourceNotFound
		}
		return errors.Join(ErrInternal, err)
	}
	return nil
}

// Incomping userID == nil ~ it's not owner
func (u *Usecase) Participate(ctx context.Context, code string, pref model.Preference, userID *string) (string, error) {
	var userUUID uuid.UUID
	if userID == nil {
		userUUID = u.resolveOwnerToken()
		userID = func(userUUID uuid.UUID) *string {
			str := userUUID.String()
			return &str
		}(userUUID)
	} else {
		userUUID, _ = uuid.Parse(*userID)
	}

	prefEmbedding, err := u.Embedder.BuildPreferenceEmbedding(ctx, pref)
	if err != nil {
		return *userID, errors.Join(ErrInternal, err)
	}

	if err := u.RoomRepository.AddPreferenceEmbedding(ctx, code, userUUID, prefEmbedding); err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return *userID, ErrResourceNotFound
		}
		return *userID, errors.Join(ErrInternal, err)
	}

	return *userID, nil
}

func (u *Usecase) ParticipantsCount(ctx context.Context, code string) (int, error) {
	count, err := u.RoomRepository.ParticipantsCount(ctx, code)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return 0, ErrResourceNotFound
		}
		return 0, errors.Join(ErrInternal, err)
	}
	return count, nil
}

func (u *Usecase) UUIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	_uuid, err := u.RoomRepository.UUIDByCode(ctx, code)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) {
			return uuid.Nil, ErrResourceNotFound
		}
		return uuid.Nil, errors.Join(err, ErrInternal)
	}

	return _uuid, err
}
