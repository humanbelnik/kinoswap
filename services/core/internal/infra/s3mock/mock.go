package s3mock

import (
	"context"
	"time"

	"github.com/humanbelnik/kinoswap/core/internal/model"
)

type S3Storage struct{}

func New() *S3Storage {
	return &S3Storage{}
}

func (s *S3Storage) Save(ctx context.Context, obj *model.Poster, readyKey *string) (string, error) {
	return "", nil
}

func (s *S3Storage) Load(ctx context.Context, readyKey string) (*model.Poster, error) {
	obj := (&model.Poster{}).NewFromData([]byte{}, "")
	return obj.(*model.Poster), nil
}

func (s *S3Storage) Update(ctx context.Context, key string, obj *model.Poster) error {
	_, err := s.Save(ctx, obj, &key)
	return err
}

func (s *S3Storage) Delete(ctx context.Context, readyKey string) error {

	return nil
}

func (s *S3Storage) GeneratePresignedURL(ctx context.Context, rawURL string, ttl time.Duration) (string, error) {
	return "", nil
}
