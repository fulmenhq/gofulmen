# gofulmen Unreleased – Go 1.26.9 floor and goneat v0.6.2

> **TODO (@3leapsdave):** Rename this file to `v0.3.7.md` or `v0.4.0.md` when choosing the next version. The `go` directive bump forces consumers onto Go 1.26.9+. `VERSION` remains `0.3.6` until that call.

Release Date: unreleased

## Overview

This draft raises the library Go floor to 1.26.9 to clear reachable standard-library vulnerabilities reported by `govulncheck` on go1.25.14 and go1.26.8. It also moves local and CI tooling to goneat v0.6.2 and digest-pins the Fulmen toolbox runner at v0.5.8.

## Key Changes

### Go 1.26.9 library floor

- `go.mod` `go` directive: 1.25.5 → 1.26.9.
- Go 1.25 is EOL.
- `govulncheck ./...` reports 9 reachable findings on go1.25.14/go1.26.8 (`net/http`, `net/textproto`, `crypto/tls`; GO-2026-6603 through GO-2026-6617) and zero on go1.26.9.

### Tooling

- goneat v0.6.0 → v0.6.2 (`Makefile` `GONEAT_VERSION`).
- CI runner `goneat-tools-runner-glibc:v0.5.6` → digest-pinned `:v0.5.8@sha256:a64b0319f4aad1069f2edd587f9d4d7bf2c52318c3c3e0c9d88650c5983c9c77`.
- External installation test Go version 1.25.14 → 1.26.9.
- GitHub Actions `actions/checkout` and `actions/setup-go` pinned to full commit SHAs.
- `.goneat/tools.yaml` Go `minimum_version` / `recommended_version`: 1.21.0 / 1.22.0 → 1.26.9.

### Platform note

goneat v0.6.2 has no `darwin/amd64` archive. Intel macOS stays on goneat v0.6.1.

## Consumer Impact

Consumers must use Go 1.26.9 or newer. No public API migration is included. Crucible remains v0.4.19.

## Testing & Validation

- `GOTOOLCHAIN=go1.26.9 go mod verify`
- `make fmt && git diff --exit-code`
- `make test`
- `make lint`
- `go build ./...`
- `govulncheck ./...`

---

For prior releases see `docs/releases/` and `RELEASE_NOTES.md`.
