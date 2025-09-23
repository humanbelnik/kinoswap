package infra_embedder_mock

import (
	"context"

	"github.com/google/uuid"
	"github.com/humanbelnik/kinoswap/core/internal/model"
)

type Embedder struct {
}

func New() *Embedder {
	return &Embedder{}
}

func (e *Embedder) EmbedMovie(ctx context.Context, ID uuid.UUID, M model.MovieMeta) error {
	return nil
}

func (e *Embedder) EmbedPreference(ctx context.Context, ID uuid.UUID, P model.Preference) error {
	return nil
}
