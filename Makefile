.PHONY: help up down restart build logs ps db-shell test clean

help:
	@echo "Available commands:"
	@echo "  make up        - Start all Docker containers (PostgreSQL & Go API)"
	@echo "  make down      - Stop all Docker containers"
	@echo "  make restart   - Restart all containers"
	@echo "  make build     - Rebuild Docker images and start containers"
	@echo "  make logs      - Tail container logs"
	@echo "  make ps        - List running container status"
	@echo "  make db-shell  - Open psql shell inside postgres container"
	@echo "  make test      - Run Go tests"
	@echo "  make clean     - Stop and remove volumes"

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

build:
	docker compose up -d --build

logs:
	docker compose logs -f

ps:
	docker compose ps

db-shell:
	docker compose exec postgres psql -U postgres -d delivery_db

test:
	go test -v ./...

clean:
	docker compose down -v
