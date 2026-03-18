.PHONY: run build test tidy fmt swagger

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
	swag init -g cmd/api/main.go -o docs
