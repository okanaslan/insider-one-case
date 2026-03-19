.PHONY: run build test tidy fmt swagger migrate migrate-status

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

swagger:
	swag init -g cmd/api/main.go -o docs --parseInternal

migrate:
	go run ./cmd/migrate up

migrate-status:
	go run ./cmd/migrate status
