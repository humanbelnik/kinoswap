package usecase_movie

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
)

var (
	ErrInvalidInput           = errors.New("invalid input")
	ErrFailedToStoreMeta      = errors.New("failed to store meta")
	ErrFailedToLoadMeta       = errors.New("failed to load meta")
	ErrFailedToUpdateMeta     = errors.New("failed to update meta")
	ErrFailedToDeleteMeta     = errors.New("failed to delete meta")
	ErrFailedToGetPreferences = errors.New("failed to get preferences")
	ErrFailedToKNN            = errors.New("failed to KNN")
	ErrFailedToReduce         = errors.New("failed to reduce")
)

type MovieRepository interface {
	Store(ctx context.Context, mm model.MovieMeta) error
	Load(ctx context.Context) ([]*model.MovieMeta, error)
	LoadByID(ctx context.Context, ID uuid.UUID) (model.MovieMeta, error)
	LoadByIDs(ctx context.Context, IDs []uuid.UUID) ([]*model.MovieMeta, error)
	Update(ctx context.Context, mm model.MovieMeta) error
	DeleteByID(ctx context.Context, ID uuid.UUID) error
	UpdateEmbedding(ctx context.Context, ID uuid.UUID, e model.Embedding) error

	KNN(ctx context.Context, k int, e model.Embedding) ([]*model.MovieMeta, error)
}

type Embedder interface {
	BuildMovieEmbedding(ctx context.Context, mm model.MovieMeta) (model.Embedding, error)
}

type EmbeddingRepository interface {
	Store(ctx context.Context, movieID uuid.UUID, e model.Embedding) error
	LoadPreferenceEmbeddings(ctx context.Context, roomID model.RoomID) ([]*model.Embedding, error)
}

type EmbeddingReducer interface {
	Reduce(ems []*model.Embedding) model.Embedding
}

type Usecase struct {
	embedder            Embedder
	repository          MovieRepository
	embeddingRepository EmbeddingRepository
	embeddingReducer    EmbeddingReducer
}

func New(
	embedder Embedder,
	meta MovieRepository,
	embeddingRepo EmbeddingRepository,
	embeddingReducer EmbeddingReducer,
) *Usecase {
	return &Usecase{
		embedder:            embedder,
		repository:          meta,
		embeddingRepository: embeddingRepo,
		embeddingReducer:    embeddingReducer,
	}
}

func (u *Usecase) KMostRelevantMovies(ctx context.Context, roomID model.RoomID, K int) ([]*model.MovieMeta, error) {
	if K <= 0 {
		return nil, fmt.Errorf("%w: K must be positive", ErrInvalidInput)
	}

	prefs, err := u.embeddingRepository.LoadPreferenceEmbeddings(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrFailedToGetPreferences, err)
	}

	reducedPref, err := func() (pref model.Embedding, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = ErrFailedToReduce
			}
		}()
		pref = u.embeddingReducer.Reduce(prefs)
		return pref, nil
	}()

	if err != nil {
		fmt.Println("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", err)
		return nil, err
	}

	mms, err := u.repository.KNN(ctx, K, reducedPref)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrFailedToKNN, err)
	}

	return mms, nil
}

func (u *Usecase) Upload(ctx context.Context, mm model.MovieMeta) error {
	if mm.Title == model.EmptyTitle {
		return ErrInvalidInput
	}

	if err := u.repository.Store(ctx, mm); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToStoreMeta, err)
	}

	var err error
	go func() {
		defer func() {
			if err != nil {
				if err := u.repository.DeleteByID(ctx, mm.ID); err != nil {
					//! Logging here
				}
			}
		}()

		ctx := context.Background()
		e, err := u.embedder.BuildMovieEmbedding(ctx, mm)
		if err != nil {
			return
		}
		if err := u.embeddingRepository.Store(ctx, mm.ID, e); err != nil {
			return
		}
	}()

	return nil
}

func (u *Usecase) DeleteMovie(ctx context.Context, id uuid.UUID) error {
	if err := u.repository.DeleteByID(ctx, id); err != nil {
		return fmt.Errorf("%w : %w", ErrFailedToDeleteMeta, err)
	}

	return nil
}

func (u *Usecase) UpdateMovie(ctx context.Context, mm model.MovieMeta) error {
	if err := u.repository.Update(ctx, mm); err != nil {
		return fmt.Errorf("%w : %w", ErrFailedToUpdateMeta, err)
	}

	go func() {
		/*
			Don't use parent HTTP context on async tasks.
			Parent context cancels when response is made.
		*/
		ctx := context.Background()
		e, err := u.embedder.BuildMovieEmbedding(ctx, mm)
		if err != nil {
			//! Logging
			return
		}
		if err := u.embeddingRepository.Store(ctx, mm.ID, e); err != nil {
			//! Logging
			return
		}
	}()
	return nil
}

func (u *Usecase) Load(ctx context.Context) ([]*model.MovieMeta, error) {
	mm, err := u.repository.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrFailedToLoadMeta, err)
	}
	return mm, nil
}
