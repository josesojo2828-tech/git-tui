APP_NAME := git-tui
BIN_DIR := bin
CMD_PATH := ./cmd/git-tui

.PHONY: all build test fmt vet run clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)
	@echo "Build complete: $(BIN_DIR)/$(APP_NAME)"

test:
	go test -race -count=1 ./...

fmt:
	gofmt -l -w .

vet:
	go vet ./...

run: build
	./$(BIN_DIR)/$(APP_NAME) log

clean:
	rm -rf $(BIN_DIR)
