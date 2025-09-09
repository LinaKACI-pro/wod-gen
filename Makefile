APP_NAME = wod-gen
MIGRATION_DIR = db/migrations
DB_URL ?= ${DB_DSN}
DOCKER_REGISTRY ?= local
TAG ?= latest
BASE_URL ?= http://localhost:8080/api/v1/wod/generate

USER ?= user123
GEN_JWT = AUTH_JWT_SECRET=$$AUTH_JWT_SECRET go run ./cmd/gen-jwt/main.go $(USER)

.PHONY: all build migrate-up migrate-down run docker-build docker-up docker-down test test-race test-k6 fmt vet lint clean gen-api tools

all: build migrate-up run

tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/kisielk/errcheck@latest
	go install github.com/daixiang0/gci@latest

migrate-up:
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" up
	@echo "Migrations applied successfully!"

migrate-down:
	migrate -path $(MIGRATION_DIR) -database "$(DB_URL)" down
	@echo "Migrations reverted successfully!"

test:
	go test -coverprofile=coverage.out -covermode=atomic ./...

test-race:
	CGO_ENABLED=1 go test -race -coverprofile=coverage.out -covermode=atomic ./...

test-k6:
	@if [ -z "$$AUTH_JWT_SECRET" ]; then \
		echo "AUTH_JWT_SECRET is not set. Run: make test-k6 AUTH_JWT_SECRET=xxxx"; \
		exit 1; \
	fi; \
	for f in tests/*.js; do \
		echo "=== Running $$f ==="; \
		TOKEN=$$($(GEN_JWT)) \
		k6 run --env BASE_URL=$(BASE_URL) --env TOKEN=$$TOKEN $$f || exit 1; \
	done

fmt:
	gofumpt -w .
	go fmt ./...
	gci write --skip-generated -s standard -s default .

vet:
	go vet ./...

lint:
	golangci-lint config verify -c golangci.yml
	golangci-lint run -c golangci.yml ./...
	command -v govulncheck >/dev/null 2>&1 && govulncheck ./... || true

docker-build:
	docker build -f build/Dockerfile -t $(APP_NAME):local .

docker-up:
	docker compose up --build

docker-down:
	docker compose down

clean:
	docker stop $(APP_NAME) postgres || true
	docker rm $(APP_NAME) postgres || true
	docker rmi $(DOCKER_REGISTRY)/$(APP_NAME):$(TAG) || true
	rm -rf bin/$(APP_NAME)

gen-api:
	cd docs && go generate ./...
