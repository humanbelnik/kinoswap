
mock-usecase-deps: mock-vote-usecase-deps mock-room-usecase-deps 

mock-vote-usecase-deps:
	mockery --dir ./internal/usecase/vote --name VoteRepository --output ./mocks/repository --filename vote.go

mock-room-usecase-deps:
	mockery --dir ./internal/usecase/room --name RoomRepository --output ./mocks/repository --filename room.go
	mockery --dir ./internal/usecase/room --name EmbeddingRepository --output ./mocks/repository --filename room_embedding.go
	mockery --dir ./internal/usecase/room --name EmptyRoomsIDSet --output ./mocks/cache --filename room.go
	mockery --dir ./internal/usecase/room --name Embedder --output ./mocks/embedder --filename room_embedder.go
