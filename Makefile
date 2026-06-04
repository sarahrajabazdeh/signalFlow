.PHONY: build test test-race lint run-api run-consumer docker-up docker-down swagger migrate

build:
	go build -o bin/api ./cmd/api/
	go build -o bin/consumer ./cmd/consumer/

test:
	go test ./...

test-race:
	go test -race ./...

lint:
	golangci-lint run

run-api:
	go run ./cmd/api/

run-consumer:
	go run ./cmd/consumer/

docker-up:
	docker compose up -d

docker-down:
	docker compose down

swagger:
	swag init -g cmd/api/main.go -o internal/api/docs

migrate:
	atlas migrate apply --url "$(DATABASE_URL)"
