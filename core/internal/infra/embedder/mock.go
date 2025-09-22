package infra_embedder_mock

import (
	"context"
)

type Embedder struct {
}

func New() *Embedder {
	return &Embedder{}
}

func (e *Embedder) Embed(ctx context.Context, v any) ([]byte, error) {
	return []byte{}, nil
}
