.PHONY: up down down-v build logs \
        migrate-up migrate-down \
        swag sqlc \
        tidy lint test test-cover \
        install-tools build-local fmt test-api

#  Docker 
up:
	docker compose up --build -d

down:
	docker compose down

down-v:
	docker compose down -v

build:
	docker compose build

logs:
	docker compose logs -f app

#  Database 
migrate-up:
	docker compose run --rm migrate

migrate-down:
	docker run --rm -v $(PWD)/migrations:/migrations \
		migrate/migrate \
		-path=/migrations \
		-database="postgres://${DB_USER}:${DB_PASSWORD}@localhost:5432/${DB_NAME}?sslmode=${DB_SSL_MODE}" \
		down 1

#  Code generation 
swag:
	swag init -g cmd/server/main.go -o docs

sqlc:
	sqlc generate

#  Go 
tidy:
	go mod tidy
fmt:
	gofmt -w .
	goimports -w .

lint:
	golangci-lint run ./...
test:
	go test ./... -v

test-cover:
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated at coverage.html"

build-local:
	go build -o bin/server ./cmd/server

test-api:
	@bash scripts/test_api.sh

#  Tools 
install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest
	go install go.uber.org/mock/mockgen@latest
	go install golang.org/x/tools/cmd/goimports@latest
	brew install golangci-lint
