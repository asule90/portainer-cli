BINARY_NAME=portainer-cli
BIN_DIR=bin
SRC=./cmd/portainer-cli

.PHONY: all build clean test install run help

all: test build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(SRC)

test:
	go test -v ./...

install: build
	@mkdir -p $(HOME)/.local/bin
	cp $(BIN_DIR)/$(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(HOME)/.local/bin"

clean:
	rm -rf $(BIN_DIR) $(BINARY_NAME)

run:
	go run $(SRC) $(ARGS)

help:
	@echo "Available targets:"
	@echo "  build    - Compile binary to $(BIN_DIR)/$(BINARY_NAME)"
	@echo "  test     - Run unit tests"
	@echo "  install  - Install binary into ~/.local/bin"
	@echo "  clean    - Remove build outputs"
	@echo "  run      - Run CLI using 'go run' (pass ARGS='...')"
