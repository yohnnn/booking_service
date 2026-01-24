.PHONY: build run test test-e2e lint clean docker-up docker-down

docker-up:
	docker compose up --build

docker-down:
	docker compose down -v

consumer:
	go run ./cmd/consumer/main.go

