
mock-usecase-deps: mock-vote-usecase-deps mock-room-usecase-deps mock-movie-usecase-deps

mock-vote-usecase-deps:
	mockery --dir ./internal/usecase/vote --name VoteRepository --output ./mocks/vote/repository --filename vote.go

mock-room-usecase-deps:
	mockery --dir ./internal/usecase/room --name RoomRepository --output ./mocks/room/repository --filename room.go
	mockery --dir ./internal/usecase/room --name EmbeddingRepository --output ./mocks/room/repository --filename embedding.go
	mockery --dir ./internal/usecase/room --name EmptyRoomsIDSet --output ./mocks/room/cache --filename room.go
	mockery --dir ./internal/usecase/room --name Embedder --output ./mocks/room/embedder --filename embedder.go

mock-movie-usecase-deps:
	mockery --dir ./internal/usecase/movie --name Embedder --output ./mocks/movie/embedder --filename embedder.go
	mockery --dir ./internal/usecase/movie --name EmbeddingRepository --output ./mocks/movie/repository --filename embedding.go
	mockery --dir ./internal/usecase/movie --name MovieRepository --output ./mocks/movie/repository --filename movie.go
	mockery --dir ./internal/usecase/movie --name PosterStorage --output ./mocks/movie/s3 --filename s3.go

ALLURE_RESULTS_DIR = allure-results
ALLURE_REPORT_DIR = allure-report

test-allure:
	rm -rf $(ALLURE_RESULTS_DIR) $(ALLURE_REPORT_DIR)
	mkdir -p $(ALLURE_RESULTS_DIR)
	
	
	go test ./... -shuffle=on -v
	
	find . -name "allure-results" -type d | while read dir; do \
		if [ "$$dir" != "./$(ALLURE_RESULTS_DIR)" ]; then \
			cp -r "$$dir"/* $(ALLURE_RESULTS_DIR)/ 2>/dev/null || true; \
		fi \
	done
	
	allure generate $(ALLURE_RESULTS_DIR) -o $(ALLURE_REPORT_DIR) --clean
