PROJECT := archive-orchestrator
BIN_DIR := bin

.PHONY: fmt vet test race build check run clean

fmt:
	gofmt -w $$(find . -name '*.go' -type f)

vet:
	go vet ./...

test:
	go test ./...

race:
	go test -race ./...

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(PROJECT) ./cmd/server

check: vet test race build

run:
	go run ./cmd/server

clean:
	rm -rf $(BIN_DIR) data
