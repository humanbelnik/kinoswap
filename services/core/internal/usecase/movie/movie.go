package usecase_movie

import (
	"context"
	"errors"
	"fmt"
	"time"

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

	ErrFailedToStorePoster    = errors.New("failed to store poster")
	ErrFailedToBuildEmbedding = errors.New("failed to build embedding")
	ErrFailedToStoreEmbedding = errors.New("failed to store embedding")
	ErrFailedToPresignURLs    = errors.New("failed to presign urls")
)

type MovieRepository interface {
	Store(ctx context.Context, mm model.MovieMeta) error
	DeleteByID(ctx context.Context, ID uuid.UUID) error

	Load(ctx context.Context) ([]*model.MovieMeta, error)
	LoadByID(ctx context.Context, ID uuid.UUID) (model.MovieMeta, error)
	LoadByIDs(ctx context.Context, IDs []uuid.UUID) ([]*model.MovieMeta, error)
	Update(ctx context.Context, mm model.MovieMeta) error

	UpdateEmbedding(ctx context.Context, ID uuid.UUID, e model.Embedding) error

	KNN(ctx context.Context, k int, e model.Embedding) ([]*model.MovieMeta, error)
}

type PosterStorage[T model.FileObject] interface {
	Save(ctx context.Context, object T, readyKey *string) (string, error)
	Delete(ctx context.Context, K string) error

	Load(ctx context.Context, K string) (T, error)

	GeneratePresignedURL(ctx context.Context, rawURL string, ttl time.Duration) (string, error)
}

type Embedder interface {
	BuildMovieEmbedding(ctx context.Context, mm model.MovieMeta) (model.Embedding, error)
}

type EmbeddingRepository interface {
	Store(ctx context.Context, movieID uuid.UUID, e model.Embedding) error
	//LoadPreferenceEmbeddings(ctx context.Context, roomID model.RoomID) ([]*model.Embedding, error)
}

type EmbeddingReducer interface {
	Reduce(ems []*model.Embedding) model.Embedding
}

type Usecase[T model.FileObject] struct {
	posterStorage PosterStorage[T]

	embedder            Embedder
	metaRepository      MovieRepository
	embeddingRepository EmbeddingRepository
	embeddingReducer    EmbeddingReducer
}

func New[T model.FileObject](
	posterStorage PosterStorage[T],
	embedder Embedder,
	meta MovieRepository,
	embeddingRepo EmbeddingRepository,
	embeddingReducer EmbeddingReducer,
) *Usecase[T] {
	return &Usecase[T]{
		posterStorage:       posterStorage,
		embedder:            embedder,
		metaRepository:      meta,
		embeddingRepository: embeddingRepo,
		embeddingReducer:    embeddingReducer,
	}
}

// func (u *Usecase[T]) KMostRelevantMovies(ctx context.Context, roomID model.RoomID, K int) ([]*model.MovieMeta, error) {
// 	if K <= 0 {
// 		return nil, fmt.Errorf("%w: K must be positive", ErrInvalidInput)
// 	}

// 	prefs, err := u.embeddingRepository.LoadPreferenceEmbeddings(ctx, roomID)
// 	if err != nil {
// 		return nil, fmt.Errorf("%w:%w", ErrFailedToGetPreferences, err)
// 	}

// 	reducedPref, err := func() (pref model.Embedding, err error) {
// 		defer func() {
// 			if r := recover(); r != nil {
// 				err = ErrFailedToReduce
// 			}
// 		}()
// 		pref = u.embeddingReducer.Reduce(prefs)
// 		return pref, nil
// 	}()

// 	if err != nil {
// 		return nil, err
// 	}

// 	mms, err := u.metaRepository.KNN(ctx, K, reducedPref)
// 	if err != nil {
// 		return nil, fmt.Errorf("%w:%w", ErrFailedToKNN, err)
// 	}

// 	return mms, nil
// }

func (u *Usecase[T]) Upload(ctx context.Context, movie model.Movie) error {
	ID := uuid.New()
	movie.MM.ID = ID
	strID := ID.String()

	if err := u.metaRepository.Store(ctx, *movie.MM); err != nil {
		return fmt.Errorf("%w: %w", ErrFailedToStoreMeta, err)
	}

	if movie.Poster != nil {
		movie.MM.PosterLink = strID
		typedPoster := any(&model.Poster{
			Filename: strID,
			Content:  movie.Poster.Content,
			MovieID:  movie.Poster.MovieID,
		}).(T)

		if _, err := u.posterStorage.Save(ctx, typedPoster, &strID); err != nil {
			_ = u.metaRepository.DeleteByID(ctx, movie.MM.ID)
			return fmt.Errorf("%w:%w", ErrFailedToStorePoster, err)
		}
	}

	emb, err := u.embedder.BuildMovieEmbedding(ctx, *movie.MM)
	if err != nil {
		_ = u.posterStorage.Delete(ctx, movie.MM.PosterLink)
		_ = u.metaRepository.DeleteByID(ctx, movie.MM.ID)
		return fmt.Errorf("%w : %w", ErrFailedToBuildEmbedding, err)
	}

	if err = u.embeddingRepository.Store(ctx, movie.MM.ID, emb); err != nil {
		_ = u.posterStorage.Delete(ctx, movie.MM.PosterLink)
		_ = u.metaRepository.DeleteByID(ctx, movie.MM.ID)
		return fmt.Errorf("%w : %w", ErrFailedToStoreEmbedding, err)
	}

	return nil
}

func (u *Usecase[T]) DeleteMovie(ctx context.Context, id uuid.UUID) error {
	_ = u.posterStorage.Delete(ctx, id.String())
	_ = u.metaRepository.DeleteByID(ctx, id)

	return nil
}

func (u *Usecase[T]) UpdateMovie(ctx context.Context, mm model.MovieMeta) error {
	if err := u.metaRepository.Update(ctx, mm); err != nil {
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

func (u *Usecase[T]) Load(ctx context.Context) ([]*model.MovieMeta, error) {
	mm, err := u.metaRepository.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w:%w", ErrFailedToLoadMeta, err)
	}

	for _, m := range mm {
		if m.PosterLink != "" {
			m.PosterLink, err = u.posterStorage.GeneratePresignedURL(ctx, m.PosterLink, 10*time.Minute)
			if err != nil {
				return nil, fmt.Errorf("%w : %w", ErrFailedToPresignURLs, err)
			}
		}
	}

	return mm, nil
}
