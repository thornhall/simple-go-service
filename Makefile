# ——— Variables —————————————————————————————————————————————
APP_NAME       := simple-service
CMD_DIR        := ./cmd/server
BIN_DIR        := ./bin
BIN_PATH       := $(BIN_DIR)/$(APP_NAME)
MIGRATE        := migrate -path db/migrations -database "$(DATABASE_URL)"
DC             := docker-compose -f docker-compose.yml
GO             := go
GOFLAGS        :=

# ——— Default target —————————————————————————————————————————
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' Makefile \
	  | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

# ——— Build & Run ——————————————————————————————————————————
.PHONY: build        ## Compile the binary
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_PATH) $(CMD_DIR)

.PHONY: run          ## Run the service (built binary)
run: build
	$(BIN_PATH)

.PHONY: dev          ## Run with hot-reload (requires Air)
dev:  ## needs github.com/cosmtrek/air installed
	air

# ——— Testing & Linting ——————————————————————————————————————
.PHONY: test         ## Run unit tests
test:
	docker compose -f docker-compose.test.yml up -d db

	DATABASE_URL="postgres://user:pass@localhost:5434/simple_service_test?sslmode=disable"
	go test ./...

	docker-compose down -v

.PHONY: fmt          ## gofmt check
fmt:
	$(GO) fmt ./...

.PHONY: vet          ## go vet
vet:
	$(GO) vet ./...

.PHONY: lint
lint:
	golangci-lint run

# ——— Database Migrations ————————————————————————————————
.PHONY: migrate-up   ## Apply all up migrations
migrate-up:  ## requires migrate CLI (golang-migrate)
	$(MIGRATE) up

.PHONY: migrate-down ## Rollback one migration
migrate-down:
	$(MIGRATE) down 1

# ——— Docker Compose ——————————————————————————————————————
.PHONY: docker-up    ## Start Postgres (docker-compose)
docker-up:
	$(DC) up -d

.PHONY: docker-down  ## Stop and remove containers
docker-down:
	$(DC) down

