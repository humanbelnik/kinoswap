
backend:
	docker-compose -f docker-compose.prod.yml up --build core-app

core-unit-tests:
	docker compose -f docker-compose.prod.yml up core-unit-tests

core-integration-tests-at-scale:
	docker-compose -f docker-compose.prod.yml up --scale core-integration-tests=3 core-integration-tests

# Helpers

sniff-traffic:
	sudo tshark -i any -Y "http and tcp.port==8080" -V

pre-commit:
	gocyclo -over 10 ./services/core
	