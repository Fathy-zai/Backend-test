.PHONY: run test build docker

run:
	env $$(cat .env | xargs) go run ./cmd/server

test:
	go test ./...

build:
	go build -o bin/weather ./cmd/server

docker:
	docker build -t weather-service:latest .
