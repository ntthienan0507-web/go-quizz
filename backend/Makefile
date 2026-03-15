.PHONY: build run test lint clean

build:
	go build -o bin/server .

run:
	go run .

test:
	go test ./... -race -count=1

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

dev:
	air
