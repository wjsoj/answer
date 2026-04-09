.PHONY: build clean ui docker docker-builder docker-push up down up-prod down-prod version

BIN := answer
DOCKER_REPO := git.pku.edu.cn/2200011523/answer
VERSION := $(shell (git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0-dev") | sed 's/^v//')
VERSION_FULL := $(shell (git describe --tags --always 2>/dev/null || echo "0.0.0-dev") | sed 's/^v//')

# Build binary locally
build:
	@CGO_ENABLED=0 GO111MODULE=on go build \
		-ldflags="-X github.com/apache/answer/cmd.Version=$(VERSION) -X 'github.com/apache/answer/cmd.Revision=$$(git rev-parse --short HEAD 2>/dev/null)'" \
		-o $(BIN) ./cmd/answer

# Clean build artifacts
clean:
	@rm -f $(BIN)

# Build UI
ui:
	@cd ui && pnpm build && cd -

BUILDX_BUILDER ?= multiarch-builder

# Ensure a multi-arch buildx builder exists
.PHONY: docker-builder
docker-builder:
	@docker buildx inspect $(BUILDX_BUILDER) > /dev/null 2>&1 || \
		docker buildx create --name $(BUILDX_BUILDER) \
			--driver docker-container \
			--platform linux/amd64,linux/arm64 \
			--use
	@docker buildx use $(BUILDX_BUILDER)

# Build local docker image (current arch only)
docker:
	@docker build \
		--build-arg GOPROXY=https://goproxy.cn,direct \
		-t $(DOCKER_REPO):$(VERSION_FULL) \
		-t $(DOCKER_REPO):latest .
	@echo "Built: $(DOCKER_REPO):$(VERSION_FULL)"

# Multi-arch build and push to registry
# Usage: make docker-push
# Requires: docker login git.pku.edu.cn
docker-push: docker-builder
	@docker buildx build \
		--builder $(BUILDX_BUILDER) \
		--platform linux/amd64,linux/arm64 \
		--build-arg GOPROXY=https://goproxy.cn,direct \
		-t $(DOCKER_REPO):$(VERSION) \
		-t $(DOCKER_REPO):$(VERSION_FULL) \
		-t $(DOCKER_REPO):latest \
		--push .
	@echo "Pushed: $(DOCKER_REPO):$(VERSION) $(DOCKER_REPO):$(VERSION_FULL) $(DOCKER_REPO):latest"

# Dev/test environment (SQLite, no external deps)
up:
	@docker compose up -d
	@echo "Dev server running at http://localhost:9080"

down:
	@docker compose down

# Production environment (PostgreSQL + MeiliSearch)
up-prod:
	@docker compose -f docker-compose.prod.yml --env-file .env.prod up -d
	@echo "Production server running at http://localhost:9080"

down-prod:
	@docker compose -f docker-compose.prod.yml --env-file .env.prod down

# Show version info
version:
	@echo "VERSION:      $(VERSION)"
	@echo "VERSION_FULL: $(VERSION_FULL)"
