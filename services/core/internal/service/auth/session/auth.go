package session_auth

import (
	"context"
	"time"

	"github.com/humanbelnik/kinoswap/core/internal/model"
)

type UserFetcher interface {
	Fetch(ctx context.Context, login string) (model.User, error)
}

type SessionCacher interface {
	Get(k string) (model.User, error)
	Set(k string, u model.User) error
	Delete(k string) error
	Exists(k string) (bool, error)
}

type Service struct {
	sessionTTL    time.Duration
	userFetcher   UserFetcher
	sessionCacher SessionCacher
}

func New(
	sessionTTL time.Duration,
	userFetcher UserFetcher,
	sessionCacher SessionCacher,
) *Service {
	return &Service{}
}
