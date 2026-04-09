# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A fork of [Apache Answer](https://github.com/apache/answer) — a Q&A forum platform with a Go backend (Gin framework) and React TypeScript frontend. Extended with MCP server support, forum sections, AI assistant features, and custom plugins.

## Build & Run Commands

### Prerequisites
- Go >= 1.24, Node.js >= 20, pnpm >= 9

### Development
```bash
make build          # Compile Go binary (output: ./answer)
make ui             # Build frontend: cd ui && pnpm build
make up             # Run dev environment with Docker (SQLite, port 9080)
make down           # Stop dev environment
make up-prod        # Run production (PostgreSQL + MeiliSearch, port 9080)
```

### Frontend Development
```bash
cd ui
pnpm install
pnpm start          # Dev server with hot reload
pnpm lint           # ESLint with auto-fix
pnpm prettier       # Format code
```

### Backend Linting
```bash
golangci-lint run    # Uses .golangci.yaml config
```

### Code Generation
```bash
# Swagger docs (from main.go annotation)
go generate ./cmd/answer/main.go
# Wire dependency injection
go generate ./cmd/wire.go
```

### Docker
```bash
make docker         # Build local Docker image
make docker-push    # Multi-arch build & push
```

## Architecture

### Backend (Go, `internal/`)

Layered architecture with Wire dependency injection:

- **`cmd/`** — Entry point. `wire.go` wires all providers. `main.go` imports plugins via blank imports.
- **`internal/controller/`** — HTTP handlers (public API). `controller_admin/` for admin endpoints.
- **`internal/service/`** — Business logic layer. Services are organized by domain (e.g., `content/`, `question_common/`, `meta_common/`).
- **`internal/repo/`** — Data access layer (GORM/XORM ORM).
- **`internal/entity/`** — Database models.
- **`internal/schema/`** — Request/response DTOs and validation.
- **`internal/router/`** — Gin route registration. `answer_api_router.go` is the main API router.
- **`internal/base/`** — Cross-cutting concerns: middleware, error handling, pagination, reason codes.

**API routes:** All public API at `/answer/api/v1/*`, admin API at `/answer/admin/api/*`.

**MCP Server:** Integrated at `/answer/api/v1/mcp` via `mcp_router.go` and `mcp_controller.go`, using `mark3labs/mcp-go`. Tool definitions in `internal/schema/mcp_tools/`.

**Plugin system:** `plugin/` defines interfaces. Plugins are imported as blank imports in `cmd/answer/main.go`. Includes search-meilisearch, storage-s3, captcha-basic, connector-basic, reviewer-glm, etc.

### Frontend (React 18 + TypeScript, `ui/src/`)

- **`router/`** — React Router 7 with lazy-loaded pages. Route definitions in `routes.ts`, path helpers in `pathFactory.ts`.
- **`pages/`** — Page components organized by feature (Questions, Admin, Users, Tags, AiAssistant, Timeline).
- **`services/`** — Axios-based API clients matching backend endpoints.
- **`stores/`** — Zustand for global state management.
- **`components/`** — Reusable UI components (Bootstrap 5 + SCSS).
- **`plugins/`** — Frontend plugin modules with independent build (`pnpm build:packages`).
- **`common/`** — Shared constants, interfaces, utilities.

Build uses react-app-rewired (CRA with custom config).

### Internationalization

Translation files in `i18n/*.yaml` (40+ languages). Backend loads them; frontend uses react-i18next. `en_US.yaml` and `zh_CN.yaml` are the primary maintained translations.

## Key Patterns

- **Dependency injection:** All service/repo/controller wiring goes through Wire in `cmd/wire.go`. After adding new providers, regenerate with `wire ./cmd/`.
- **Error handling:** Backend uses reason codes defined in `internal/base/reason/reason.go` mapped to i18n keys.
- **Database:** SQLite for dev, PostgreSQL for production. ORM entities use XORM tags.
- **Go module path:** `github.com/apache/answer` (retained from upstream).
