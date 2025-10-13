package infra_embedder

import (
	"context"

	"github.com/humanbelnik/kinoswap/core/gen/proto"
	"github.com/humanbelnik/kinoswap/core/internal/config"
	"github.com/humanbelnik/kinoswap/core/internal/model"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Embedder struct {
	conn   *grpc.ClientConn
	client proto.EmbeddingServiceClient
}

func MustEstablishConnection(cfg config.Embedder) *Embedder {
	conn, err := grpc.NewClient(cfg.Host+":"+cfg.Port, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return &Embedder{
		conn:   conn,
		client: proto.NewEmbeddingServiceClient(conn),
	}
}

func (e *Embedder) Close() error {
	return e.conn.Close()
}

func (e *Embedder) BuildPreferenceEmbedding(ctx context.Context, P model.Preference) (model.Embedding, error) {
	req := &proto.PreferenceEmbeddingRequest{
		Text: P.Text,
	}

	resp, err := e.client.CreatePreferenceEmbedding(ctx, req)
	return model.Embedding(resp.GetEmbedding()), err
}

func (e *Embedder) BuildMovieEmbedding(ctx context.Context, mm model.MovieMeta) (model.Embedding, error) {
	req := &proto.MovieEmbeddingRequest{
		Title:    mm.Title,
		Overview: mm.Overview,
		Year:     int32(mm.Year),
		Rating:   float32(mm.Rating),
		Genres:   mm.Genres,
	}

	resp, err := e.client.CreateMovieEmbedding(ctx, req)
	return model.Embedding(resp.GetEmbedding()), err

}
