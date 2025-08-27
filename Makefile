
BINARY_NAME=yamlforge

GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

BUILD_DIR=build
CMD_DIR=cmd/yamlforge

.PHONY: all
all: build

.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) $(CMD_DIR)/main.go

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

.PHONY: test
test:
	$(GOTEST) -v ./...

.PHONY: deps
deps:
	$(GOGET) -v ./...

.PHONY: run-blog
run-blog: build
	./$(BINARY_NAME) serve examples/blog.yaml

.PHONY: run-tasks
run-tasks: build
	./$(BINARY_NAME) serve examples/tasks.yaml

.PHONY: validate
validate: build
	./$(BINARY_NAME) validate examples/blog.yaml
	./$(BINARY_NAME) validate examples/tasks.yaml

.PHONY: dev
dev: build run-blog

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make build        - Build the yamlforge binary"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make test         - Run all tests"
	@echo "  make deps         - Download dependencies"
	@echo "  make run-blog     - Run the blog example"
	@echo "  make run-tasks    - Run the tasks example"
	@echo "  make validate     - Validate all example YAML files"
	@echo "  make dev          - Build and run blog example (development)"
	@echo "  make help         - Show this help message"

.DEFAULT_GOAL := help
