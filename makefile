
backend:
	docker-compose -f docker-compose.prod.yml up --build core-app

core-unit-tests:
	docker compose -f docker-compose.prod.yml up core-unit-tests

core-integration-tests:
	docker compose -f docker-compose.prod.yml up core-integration-tests

sniff-traffic:
	sudo tshark -i any -Y "http and tcp.port==8080" -V