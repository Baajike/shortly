.PHONY: help setup up down logs \
        backend frontend \
        install-backend install-frontend \
        test-backend lint-backend build-backend build-frontend

help:
	@echo "Shortly — available targets:"
	@echo "  setup             Install all dependencies (run once after clone)"
	@echo "  up                Start Postgres + Redis via Docker Compose"
	@echo "  down              Stop Docker Compose services"
	@echo "  logs              Tail Docker Compose logs"
	@echo "  backend           Run backend with live-reload (requires air)"
	@echo "  frontend          Run frontend dev server"
	@echo "  build-backend     Compile backend binary to backend/bin/server"
	@echo "  build-frontend    Bundle frontend to frontend/dist/"
	@echo "  test-backend      Run Go tests with race detector"
	@echo "  lint-backend      Run golangci-lint on backend"

setup: install-backend install-frontend
	@echo "Setup complete. Copy .env.example to .env and configure it."

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

backend:
	cd backend && air

frontend:
	cd frontend && npm run dev

install-backend:
	cd backend && go mod download && go mod tidy

install-frontend:
	cd frontend && npm install

build-backend:
	cd backend && CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/server ./cmd/server

build-frontend:
	cd frontend && npm run build

test-backend:
	cd backend && go test ./... -v -race -coverprofile=coverage.out
	cd backend && go tool cover -html=coverage.out -o coverage.html

lint-backend:
	cd backend && golangci-lint run ./...
