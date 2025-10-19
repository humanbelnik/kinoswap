package integrationtest

import (
	"context"
	"testing"

	infra_pg_init "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/init"
	infra_postgres_room "github.com/humanbelnik/kinoswap/core/internal/infra/postgres/room"
	room_usecase "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
	mocks_embedder "github.com/humanbelnik/kinoswap/core/internal/usecase/room/mocks/room/embedder"

	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/assert"
)

type UsecaseRoomIntegrationSuite struct {
	suite.Suite
	uc *room_usecase.Usecase
}

func initRoomUsecase(t provider.T) *room_usecase.Usecase {
	cfg := getConfig()

	pgConn := infra_pg_init.MustEstablishConn(cfg.Postgres)
	roomRepository := infra_postgres_room.New(pgConn)
	embedder := mocks_embedder.NewEmbedder(t)

	usecase := room_usecase.New(roomRepository, embedder)
	return usecase
}

func (s *UsecaseRoomIntegrationSuite) BeforeAll(t provider.T) {
	s.uc = initRoomUsecase(t)
}

func (s *UsecaseRoomIntegrationSuite) TestIntegrationBook(t provider.T) {
	ctx := context.Background()

	tt := []struct {
		name     string
		setup    func()
		teardown func(roomCode string)
	}{
		{
			name: "successfully book room",
			setup: func() {
			},
			teardown: func(roomCode string) {
				ctx := context.Background()
				s.uc.Free(ctx, roomCode)
			},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t provider.T) {
			test.setup()

			roomCode, ownerToken, err := s.uc.Book(ctx)

			assert.NoError(t, err)
			assert.NotEmpty(t, roomCode)
			assert.NotEmpty(t, ownerToken)
			assert.Len(t, roomCode, 6)

			status, err := s.uc.Status(ctx, roomCode)
			assert.NoError(t, err)
			assert.Equal(t, "LOBBY", status)

			isOwner, err := s.uc.IsOwner(ctx, roomCode, ownerToken)
			assert.NoError(t, err)
			assert.True(t, isOwner)

			test.teardown(roomCode)
		})
	}
}

func TestRoomIntegrationSuite(t *testing.T) {
	suite.RunSuite(t, new(UsecaseRoomIntegrationSuite))
}
