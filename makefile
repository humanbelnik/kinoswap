
backend:
	docker compose -f docker-compose.prod.yml up core-migrate
	docker exec postgres-core-service cp /tmp/pg_hba.conf /var/lib/postgresql/data/pg_hba.conf
	docker exec postgres-core-service psql -U admin -d test -c 'SELECT pg_reload_conf();'   
	docker compose -f docker-compose.prod.yml up -d core-postgres-replica
	docker-compose -f docker-compose.prod.yml up --build core-app core-app-ro1 core-app-ro2 
	

core-unit-tests:
	docker compose -f docker-compose.prod.yml up core-unit-tests

core-integration-tests-at-scale:
	docker-compose -f docker-compose.prod.yml up --scale core-integration-tests=3 core-integration-tests

# Helpers

sniff-traffic:
	sudo tshark -i any -Y "http and tcp.port==8080" -V
