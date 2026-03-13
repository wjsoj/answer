.PHONY: build clean ui docker docker-multiarch version

BIN=answer
DIR_SRC=./cmd/answer
DOCKER_CMD=docker

GO_ENV=CGO_ENABLED=0 GO111MODULE=on
Revision=$(shell git rev-parse --short HEAD 2>/dev/null || echo "")
# Derive version from git tag: v2.1.0 -> 2.1.0, fallback to 0.0.0-dev
VERSION=$(shell (git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0-dev") | sed 's/^v//')
# Full version with commit info: v2.1.0-3-g1a2b3c4 -> 2.1.0-3-g1a2b3c4, or just commit hash
VERSION_FULL=$(shell (git describe --tags --always 2>/dev/null || echo "0.0.0-dev") | sed 's/^v//')
GO_FLAGS=-ldflags="-X github.com/apache/answer/cmd.Version=$(VERSION) -X 'github.com/apache/answer/cmd.Revision=$(Revision)' -X 'github.com/apache/answer/cmd.Time=`date +%s`' -extldflags -static"
GO=$(GO_ENV) "$(shell which go)"

GOLANGCI_VERSION ?= v2.6.2
TOOLS_BIN := $(shell mkdir -p build/tools && realpath build/tools)

GOLANGCI = $(TOOLS_BIN)/golangci-lint-$(GOLANGCI_VERSION)
$(GOLANGCI):
	rm -f $(TOOLS_BIN)/golangci-lint*
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/$(GOLANGCI_VERSION)/install.sh | sh -s -- -b $(TOOLS_BIN) $(GOLANGCI_VERSION)
	mv $(TOOLS_BIN)/golangci-lint $(TOOLS_BIN)/golangci-lint-$(GOLANGCI_VERSION)

build: generate
	@$(GO) build $(GO_FLAGS) -o $(BIN) $(DIR_SRC)

# https://dev.to/thewraven/universal-macos-binaries-with-go-1-16-3mm3
universal: generate
	@GOOS=darwin GOARCH=amd64 $(GO_ENV) $(GO) build $(GO_FLAGS) -o ${BIN}_amd64 $(DIR_SRC)
	@GOOS=darwin GOARCH=arm64 $(GO_ENV) $(GO) build $(GO_FLAGS) -o ${BIN}_arm64 $(DIR_SRC)
	@lipo -create -output ${BIN} ${BIN}_amd64 ${BIN}_arm64
	@rm -f ${BIN}_amd64 ${BIN}_arm64

generate:
	@$(GO) get github.com/swaggo/swag/cmd/swag@v1.16.3
	@$(GO) get github.com/google/wire/cmd/wire@v0.5.0
	@$(GO) get go.uber.org/mock/mockgen@v0.6.0
	@$(GO) install github.com/swaggo/swag/cmd/swag@v1.16.3
	@$(GO) install github.com/google/wire/cmd/wire@v0.5.0
	@$(GO) install go.uber.org/mock/mockgen@v0.6.0
	@$(GO) generate ./...
	@$(GO) mod tidy

check:
	@mockgen -version
	@swag -v
	@wire flags

test:
	@$(GO) test ./internal/repo/repo_test

# clean all build result
clean:
	@$(GO) clean ./...
	@rm -f $(BIN)

install-ui-packages:
	@corepack enable
	@corepack prepare pnpm@9.7.0 --activate

ui:
	@cd ui && pnpm pre-install && pnpm build && cd -

lint: generate $(GOLANGCI)
	@bash ./script/check-asf-header.sh
	$(GOLANGCI) run

lint-fix: generate $(GOLANGCI)
	@bash ./script/check-asf-header.sh
	$(GOLANGCI) run --fix

all: clean build

DOCKER_REPO=git.pku.edu.cn/2200011523/answer

version:
	@echo "VERSION:      $(VERSION)"
	@echo "VERSION_FULL: $(VERSION_FULL)"
	@echo "REVISION:     $(Revision)"

# Build Docker image for current architecture (local use)
docker:
	$(DOCKER_CMD) build \
		--build-arg GOPROXY=https://goproxy.cn,direct \
		-t $(DOCKER_REPO):$(VERSION_FULL) \
		-t $(DOCKER_REPO):latest .
	@echo "Built: $(DOCKER_REPO):$(VERSION_FULL)"

# Build and push multi-arch Docker images (amd64 + arm64)
docker-multiarch:
	$(DOCKER_CMD) buildx build \
		--platform linux/amd64,linux/arm64 \
		--build-arg GOPROXY=https://goproxy.cn,direct \
		-t $(DOCKER_REPO):$(VERSION) \
		-t $(DOCKER_REPO):$(VERSION_FULL) \
		-t $(DOCKER_REPO):latest \
		--push .
	@echo "Pushed: $(DOCKER_REPO):$(VERSION) $(DOCKER_REPO):$(VERSION_FULL) $(DOCKER_REPO):latest"
