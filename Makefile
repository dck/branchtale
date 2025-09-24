.PHONY: build test clean install help

all: test build

build:
	go build -o ./bin/branchtale ./cmd/branchtale

test:
	go test ./...

clean:
	rm -f ./bin/branchtale coverage.out coverage.html

install:
	go install ./cmd/branchtale

tidy:
	go mod tidy

help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  test          - Run tests"
	@echo "  clean         - Clean build artifacts"
	@echo "  install       - Install the binary"
	@echo "  tidy          - Tidy dependencies"
	@echo "  help          - Show this help message"
