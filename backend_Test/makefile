# TOM: I had a look and it doesn't look like make files are common for go. Probably the reason why is that it's already pretty easy to get tests and your main app running with go run and go test. That said, this approach works, but we usually use bash scripts to handle more complex stuff at Zai
.PHONY: run test build docker

run:
	env $$(cat .env | xargs) go run ./cmd/server

test:
	go test ./...

build:
	go build -o bin/weather ./cmd/server

docker:
	docker build -t weather-service:latest .
