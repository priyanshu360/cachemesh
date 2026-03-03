.PHONY: build run test clean docker-build docker-run help

BINARY_NAME=cachemesh
PORT=8080

help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  run          - Run the server"
	@echo "  test         - Run tests"
	@echo "  clean        - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"

build:
	go build -o $(BINARY_NAME) .

run:
	go run . -config=config.yaml

run-port:
	go run . -port=$(PORT)

test:
	go test -v ./...

clean:
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME)-linux-amd64

docker-build:
	docker build -t $(BINARY_NAME):latest .

docker-run:
	docker run -p $(PORT):8080 $(BINARY_NAME):latest

docker-buildx:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(BINARY_NAME):latest .
