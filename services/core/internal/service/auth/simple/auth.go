package servie_simple_auth

//! Dumbest auth

import (
	"errors"
	"os"
	"time"

	"github.com/google/uuid"
)

type Token = string

var (
	ErrInternal  = errors.New("internal error")
	ErrWrongCode = errors.New("wrong code")
)

type SessionCache interface {
	Set(key string, value string, ttl time.Duration) error
	Get(key string) (string, error)
}

type Service struct {
	secret       string
	sessionCache SessionCache
	ttl          time.Duration
}

func New(
	secret *string,
	sessionCache SessionCache,
	ttl *time.Duration,
) *Service {
	if ttl == nil {
		ttl = func() *time.Duration {
			defaultTokenTTL := time.Minute * 10
			return &defaultTokenTTL
		}()
	}

	if secret == nil {
		secret = func() *string {
			secret := os.Getenv("ADMIN_SECRET")
			return &secret
		}()
		if *secret == "" {
			*secret = "shared"
		}
	}

	return &Service{
		secret:       *secret,
		sessionCache: sessionCache,
		ttl:          *ttl,
	}
}

func (s *Service) Auth(code string) (string, error) {
	const activeSession = "active"

	if code != s.secret {
		return "", ErrWrongCode
	}

	t := s.genToken()
	if err := s.sessionCache.Set(t, activeSession, s.ttl); err != nil {
		return "", errors.Join(ErrInternal, err)
	}

	return t, nil
}

func (s *Service) IsValid(t string) (bool, error) {
	v, err := s.sessionCache.Get(t)
	if err != nil {
		return false, errors.Join(ErrInternal, err)
	}

	return v != "", nil
}

func (s *Service) genToken() string {
	return uuid.New().String()
}
