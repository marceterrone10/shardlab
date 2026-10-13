GO ?= go
BIN_DIR ?= bin
SIZE ?= 100k
SEED ?= 42

.PHONY: help fmt vet test build seed up down clean

help:
	@echo "Targets:"
	@echo "  make build             Build coordinator, analyzer and seed binaries"
	@echo "  make test              Run the Go test suite"
	@echo "  make seed SIZE=100k    Generate deterministic CRM CSV fixtures"
	@echo "  make up                Start baseline, three shards and coordinator"
	@echo "  make down              Stop the local stack"

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

build:
	mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/shardlab-coordinator ./coordinator/cmd/shardlab-coordinator
	$(GO) build -o $(BIN_DIR)/shardlab-analyzer ./analyzer/cmd/shardlab-analyzer
	$(GO) build -o $(BIN_DIR)/shardlab-seed ./seed/cmd/shardlab-seed

seed:
	$(GO) run ./seed/cmd/shardlab-seed \
		-size "$(SIZE)" \
		-seed "$(SEED)" \
		-output .data/seed

up:
	docker compose up --build -d

down:
	docker compose down

clean:
	rm -rf $(BIN_DIR) .data
