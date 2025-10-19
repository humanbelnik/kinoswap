package usecase_movie

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
)

// ! Convention
// ! Return ErrResourceNotFound if there're no such resource.
// ! ex. don't return empty collections on queries with error equal to nil. Return apropriate error instead
var (
	ErrInternal                  = errors.New("internal error")
	ErrResourceNotFound          = errors.New("no such resource")
	ErrInvalidEmbeddingDimension = errors.New("invalid embedding dimension")
)

//go:generate mockery --name=MetaRepository --output=./mocks/movie/repository --filename=meta_repository.go
type MetaRepository interface {
	Store(ctx context.Context, mm model.MovieMeta) error
	Delete(ectx context.Context, id uuid.UUID) error
	LoadAll(ctx context.Context) ([]*model.MovieMeta, error)
	LoadSome(ctx context.Context, ids []uuid.UUID) ([]*model.MovieMeta, error)
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	StoreEmbedding(ctx context.Context, id uuid.UUID, e model.Embedding) error
}

//go:generate mockery --name=Embedder --output=./mocks/movie/embedder --filename=embedder.go
type Embedder interface {
	BuildMovieEmbedding(ctx context.Context, mm model.MovieMeta) (model.Embedding, error)
}

//go:generate mockery --name=EmbeddingReducer --output=./mocks/movie/embedder --filename=embedding_reducer.go
type EmbeddingReducer interface {
	Reduce(ems []*model.Embedding) model.Embedding
}

//go:generate mockery --name=PosterRepository --output=./mocks/movie/repository --filename=poster_repository.go
type PosterRepository interface {
	Save(ctx context.Context, object *model.Poster, readyKey *string) (string, error)
	Delete(ctx context.Context, key string) error
	Load(ctx context.Context, key string) (*model.Poster, error)
	GeneratePresignedURL(ctx context.Context, rawURL string, ttl time.Duration) (string, error)
}

type Usecase struct {
	MetaRepository   MetaRepository
	PosterRepository PosterRepository
	Embedder         Embedder
	EmbeddingReducer EmbeddingReducer
}

func New(
	metaRepository MetaRepository,
	posterRepository PosterRepository,
	embedder Embedder,
	embeddingReducer EmbeddingReducer,
) *Usecase {
	return &Usecase{
		MetaRepository:   metaRepository,
		PosterRepository: posterRepository,
		Embedder:         embedder,
		EmbeddingReducer: embeddingReducer,
	}
}

// Implement Functional builder patter for Movie upload.
// - Store meta
// - Store Poster file (if specified)
// - Get embedding from external embedder
// - Store embedding in meta
// If something goes wrong, rollback to the beginning state

type Op struct {
	exec     func(ctx context.Context) error
	rollback func(ctx context.Context) error
}

type MovieBuilder struct {
	ops   []Op
	uc    *Usecase
	movie model.Movie
}

func NewMovieBuilder(u *Usecase, movie model.Movie) *MovieBuilder {
	if movie.MM.ID == uuid.Nil {
		movie.MM.ID = uuid.New()
	}

	return &MovieBuilder{
		movie: movie,
		uc:    u,
	}
}

func (b *MovieBuilder) WithMeta(ctx context.Context) *MovieBuilder {
	op := Op{
		exec: func(ctx context.Context) error {
			err := b.uc.MetaRepository.Store(ctx, *b.movie.MM)
			if err != nil {
				return errors.Join(ErrInternal, err)
			}
			return nil
		},
		rollback: func(ctx context.Context) error {
			return b.uc.MetaRepository.Delete(ctx, b.movie.MM.ID)
		},
	}
	b.ops = append(b.ops, op)
	return b
}

func (b *MovieBuilder) WithPoster(ctx context.Context) *MovieBuilder {
	if b.movie.Poster == nil {
		return b
	}

	strID := b.movie.MM.ID.String()
	b.movie.MM.PosterLink = strID

	op := Op{
		exec: func(ctx context.Context) error {
			_, err := b.uc.PosterRepository.Save(ctx, &model.Poster{
				Filename: strID,
				Content:  b.movie.Poster.Content,
				MovieID:  b.movie.Poster.MovieID,
			}, &strID)
			if err != nil {
				return errors.Join(ErrInternal, err)
			}
			return nil
		},
		rollback: func(ctx context.Context) error {
			return b.uc.PosterRepository.Delete(ctx, b.movie.MM.PosterLink)
		},
	}
	b.ops = append(b.ops, op)
	return b
}

func (b *MovieBuilder) WithEmbedding(ctx context.Context) *MovieBuilder {
	op := Op{
		exec: func(ctx context.Context) error {
			emb, err := b.uc.Embedder.BuildMovieEmbedding(ctx, *b.movie.MM)
			if err != nil {
				return errors.Join(ErrInternal, err)
			}
			if len(emb) != model.EmbeddingDimension {
				return ErrInvalidEmbeddingDimension
			}
			err = b.uc.MetaRepository.StoreEmbedding(ctx, b.movie.MM.ID, emb)
			if err != nil {
				return errors.Join(ErrInternal, err)
			}
			return nil
		},
		rollback: func(ctx context.Context) error {
			return nil
		},
	}
	b.ops = append(b.ops, op)
	return b
}

func (b *MovieBuilder) Execute(ctx context.Context) error {
	var successfullyCompletedOps int

	for _, op := range b.ops {
		if err := op.exec(ctx); err != nil {
			b.rollback(ctx, successfullyCompletedOps)
			return err
		}
		successfullyCompletedOps++
	}

	return nil
}

func (b *MovieBuilder) rollback(ctx context.Context, completedOpsCount int) {
	for i := completedOpsCount; i > -1; i-- {
		if b.ops[i].rollback != nil {
			if err := b.ops[i].rollback(ctx); err != nil {
			}
		}
	}
}

func (u *Usecase) Upload(ctx context.Context, movie model.Movie) error {
	return NewMovieBuilder(u, movie).
		WithMeta(ctx).
		WithPoster(ctx).
		WithEmbedding(ctx).
		Execute(ctx)
}

func (u *Usecase) Delete(ctx context.Context, id uuid.UUID) error {
	exists, err := u.Exists(ctx, id)
	if err != nil {
		return errors.Join(ErrInternal, err)
	}

	if !exists {
		return ErrResourceNotFound
	}

	if err := u.MetaRepository.Delete(ctx, id); err != nil {
		return errors.Join(ErrInternal, err)
	}

	if err := u.PosterRepository.Delete(ctx, id.String()); err != nil {
		return errors.Join(ErrInternal, err)
	}

	return nil
}

func (u *Usecase) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	exists, err := u.MetaRepository.Exists(ctx, id)
	if err != nil {
		return false, errors.Join(ErrInternal, err)
	}
	return exists, nil
}

// Giving presigned URLs to access files from S3 directly from the client
func (u *Usecase) LoadAll(ctx context.Context) ([]*model.MovieMeta, error) {
	mm, err := u.MetaRepository.LoadAll(ctx)
	if err != nil {
		return nil, errors.Join(ErrInternal, err)
	}

	if len(mm) == 0 {
		return nil, ErrResourceNotFound
	}

	for _, m := range mm {
		if m.PosterLink != "" {
			m.PosterLink, err = u.PosterRepository.GeneratePresignedURL(ctx, m.PosterLink, 10*time.Minute)
			if err != nil {
				return nil, errors.Join(ErrInternal, err)
			}
		}
	}

	return mm, nil
}

func (u *Usecase) LoadSome(ctx context.Context, ids []uuid.UUID) ([]*model.MovieMeta, error) {
	mm, err := u.MetaRepository.LoadSome(ctx, ids)
	if err != nil {
		return nil, errors.Join(ErrInternal, err)
	}

	if len(mm) == 0 {
		return nil, ErrResourceNotFound
	}

	for _, m := range mm {
		if m.PosterLink != "" {
			m.PosterLink, err = u.PosterRepository.GeneratePresignedURL(ctx, m.PosterLink, 10*time.Minute)
			if err != nil {
				return nil, errors.Join(ErrInternal, err)
			}
		}
	}

	return mm, nil
}
