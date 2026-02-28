.PHONY: build run test lint docker-up docker-down consumer 

build:
	go build -o bin/server ./cmd/server
	go build -o bin/consumer ./cmd/consumer

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v

consumer:
	go run ./cmd/consumer/main.go

test:
	go test -v ./internal/service/... -count=1

lint:
	golangci-lint run



