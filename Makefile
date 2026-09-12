.PHONY: run build migrate-up migrate-down docker-up docker-down test lint tidy

APP_NAME    = studio-api
BUILD_DIR   = ./bin
MIGRATE_URL = postgres://studio:studio@localhost:35433/studio_db?sslmode=disable
MIGRATE_DIR = ./migrations

run:
	go run ./cmd/main.go

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/main.go

migrate-up:
	migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_URL)" up

migrate-down:
	migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_URL)" down 1

migrate-force:
	migrate -path $(MIGRATE_DIR) -database "$(MIGRATE_URL)" force $(VERSION)

docker-up:
	docker compose up -d

docker-down:
	docker compose down

test:
	go test ./... -v -race -timeout 60s

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
