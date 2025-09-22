package infra_qdrant_mock

import (
	"context"

	"github.com/humanbelnik/kinoswap/core/internal/model"
)

type Repository struct{}

func New() *Repository {
	return &Repository{}
}

func (r *Repository) Store(ctx context.Context, ID model.EID, E model.Embedding) error {
	return nil
}
func (r *Repository) Load(ctx context.Context, ID model.EID) (model.Embedding, error) {
	return model.Embedding{}, nil
}
func (r *Repository) Delete(ctx context.Context, ID model.EID) error {
	return nil
}
func (r *Repository) KNN(ctx context.Context, K int, E model.Embedding) ([]model.EID, error) {
	return []model.EID{}, nil
}
