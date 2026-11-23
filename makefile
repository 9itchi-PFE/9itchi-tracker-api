.PHONY: build test lint run docker

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run

run:
	go run main.go

docker:
	docker build -t 9itchi-tracker-api .
