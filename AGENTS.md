# AGENTS.md — ai-librarian

MuxCore sidecar module (`ai-librarian`).

## Module identity

| Field | Value |
|-------|-------|
| Directory | `ai-librarian` |
| Role | `ai` |
| Capability | `ai.librarian` |

## Build

```bash
cd ai-librarian
go test ./...
make lint
```

## Agent rules

- Modules run as gRPC sidecars; capabilities are the security boundary.
- TLS required in production (`MUXCORE_INSECURE_DISABLE_TLS` is dev-only).
- Match existing Go patterns; run `gofmt` and package tests before finishing.
