# Load environment variables from .env if the file exists
include .env.exemple
export $(shell sed 's/=.*//' .env.exemple)

APP_NAME = wod-gen
MIGRATION_DIR = db/migrations
DOCKER_REGISTRY ?= local
TAG ?= latest
LATEST  := latest
BASE_URL ?= http://localhost:8080/api/v1/wod/generate
TOKEN ?= eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIiwiZXhwIjoxNzU3NDM3NDY4LCJpYXQiOjE3NTczNTEwNjh9.FsE9SDr-StQIGN0b3aOO-89VF8-cI-ohydROaqsd6zM

.PHONY: all build migrate-up migrate-down run run-binary docker-build docker-push docker-run-postgres docker-run-app docker-run docker-compose-dev docker-compose-prod test-k6 clean

all: build migrate-up run

tools:
	@echo "==> Installing tools"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(LATEST)
	@go install mvdan.cc/gofumpt@$(LATEST)
	@go install golang.org/x/vuln/cmd/govulncheck@$(LATEST)
	@go install github.com/kisielk/errcheck@$(LATEST)
	@go install mvdan.cc/gofumpt@$(LATEST)
	@go install github.com/daixiang0/gci@latest

migrate-up:
	goose -dir ${MIGRATION_DIR} postgres "${DB_DSN}" up
	@echo "Migrations applied successfully!"

migrate-down:
	goose -dir ${MIGRATION_DIR} postgres "${DB_DSN}" down
	@echo "Migrations reverted successfully!"

test:
	go test -coverprofile=coverage.out -covermode=atomic ./...

test-race:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

test-k6:
	@for f in tests/*.js; do \
		echo "=== Running $$f ==="; \
		k6 run --env BASE_URL=$(BASE_URL) --env TOKEN=$(TOKEN) $$f || exit 1; \
	done

fmt:
	gofumpt -w .
	go fmt ./...
	gci write --skip-generated -s standard -s default .

vet:
	go vet ./...

lint:
	@echo "-> golangci-lint"
	@golangci-lint config verify -c golangci.yml
	@golangci-lint run -c golangci.yml ./...
	@echo "-> govulncheck"; command -v govulncheck >/dev/null 2>&1 && govulncheck ./... || true

# Docker targets
docker-build:
	docker build -f build/Dockerfile -t wod-gen:local .

docker-run:
	docker run --rm -p 8080:8080 \
		-e PORT=8080 \
		-e API_KEYS=devkey \
		-e RATE_LIMIT_STRATEGY=token \
		-e ENV=dev \
		wod-gen:local

docker-up:
	docker compose up --build

docker-down:
	docker compose down

# Clean up
clean:
	@echo "Stopping and removing containers..."
	docker stop $(APP_NAME) user-service-postgres || true
	docker rm $(APP_NAME) user-service-postgres || true
	@echo "Removing Docker image..."
	docker rmi $(DOCKER_REGISTRY)/$(APP_NAME):$(TAG) || true
	@echo "Removing built binary..."
	rm -rf bin/$(APP_NAME)
	@echo "Cleanup completed!"

gen-api:
	cd docs
	go generate ./...
