# AGENTS.md

## Cursor Cloud specific instructions

GoFast is a single-process Go web framework. There is no Docker, Makefile, or frontend build step.

### Prerequisites

- Go **1.25+** (see `go.mod`)
- Config: `config/config.yaml` (committed)

### First-time setup (after clone)

Before the first `go run` or `fast` CLI command, create writable directories (SQLite fails with "out of memory" if `database/` is missing):

```bash
mkdir -p database storage/logs storage/app
go run . fast db:migrate   # optional but recommended
go run . fast db:seed      # optional test data (Alice/Bob)
```

### Running the HTTP server

```bash
go run main.go
# or: go run .
```

Server listens on **http://0.0.0.0:3000** (`server.port` in config).

### Fast CLI

```bash
go run . fast list
go run . fast db:migrate
go run . fast db:seed
```

### Tests and lint

```bash
go test ./...
go vet ./...
```

There is no `golangci-lint` config in this repo.

### Key endpoints (smoke test)

| Endpoint | Purpose |
|----------|---------|
| `GET /health` | Health check → `{"status":"ok"}` |
| `GET /api/ping` | Public API → `{"message":"pong"}` |
| `GET /admin/users/` | Admin user list (needs `Authorization: Bearer <any>` stub) |
| `GET /pages/dashboard` | Admin HTML dashboard |

### Notes

- Default DB is **SQLite** at `database/gofast.db` (gitignored).
- gRPC is defined in config but **disabled** in `main.go` (commented `RegisterGRPC` / `GRPC().Run()`).
- Auth middlewares are stubs; any Bearer token works for admin routes.
- Admin API paths use trailing slashes (e.g. `/admin/users/`).
