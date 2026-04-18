BACKEND_DIR := backend
WEB_DIR     := web
DOCKER_DIR  := docker
MIGRATE     := migrate
DB_URL      := postgres://$(shell grep DB_USER $(BACKEND_DIR)/.env | cut -d= -f2):$(shell grep DB_PASSWORD $(BACKEND_DIR)/.env | cut -d= -f2)@$(shell grep DB_HOST $(BACKEND_DIR)/.env | cut -d= -f2):$(shell grep DB_PORT $(BACKEND_DIR)/.env | cut -d= -f2)/$(shell grep DB_NAME $(BACKEND_DIR)/.env | cut -d= -f2)?sslmode=$(shell grep DB_SSLMODE $(BACKEND_DIR)/.env | cut -d= -f2)

.PHONY: dev-be dev-web dev-db migrate-up migrate-down migrate-reset build-be build-web tidy lint

# ── Development ───────────────────────────────────────────────────────────────

dev-be:
	cd $(BACKEND_DIR) && air

dev-web:
	cd $(WEB_DIR) && pnpm dev

dev-db:
	cd $(DOCKER_DIR) && docker compose up db -d

# ── Database ──────────────────────────────────────────────────────────────────

migrate-up:
	$(MIGRATE) -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" down 1

migrate-reset:
	$(MIGRATE) -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" drop -f
	$(MIGRATE) -path $(BACKEND_DIR)/migrations -database "$(DB_URL)" up

# ── Build ─────────────────────────────────────────────────────────────────────

build-be:
	cd $(BACKEND_DIR) && go build -o bin/api ./cmd/api

build-web:
	cd $(WEB_DIR) && pnpm build

# ── Utility ───────────────────────────────────────────────────────────────────

tidy:
	cd $(BACKEND_DIR) && go mod tidy

lint:
	cd $(BACKEND_DIR) && golangci-lint run ./...
