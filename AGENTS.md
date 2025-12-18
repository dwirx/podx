# Repository Guidelines

## Project Structure & Module Organization
- CLI entrypoint lives in `main.go`; core packages sit under `crypto/` (encryption backends), `project/` (PODX config + secret orchestration), `parser/` (ENV parsing/rewriting), `keygen/` (Age key helpers), and `updater/` (self-update logic).  
- Installation helpers are in `install.sh` and `install.ps1`; release/cross-build recipes in `Makefile`.  
- Binary artifacts land in the repo root as `podx` or platform-specific `podx-<os>-<arch>`; config defaults to `.podx.yaml`, secrets typically `.env` with encrypted siblings ending in `.podx`.

## Build, Test, and Development Commands
- `go build -o podx .` — local build with injected `Version`/`BuildTime` via ldflags handled in the `Makefile`.  
- `make build` / `make install` — build and optionally copy the binary into `/usr/local/bin`.  
- `make release` — cross-compile all targets and emit checksums (used for GitHub Releases).  
- `./podx <command>` — run the CLI (e.g., `./podx encrypt-all`, `./podx env decrypt -i .env.podx -o .env`).  
- `go test ./...` — add unit tests next to the packages they cover; currently none exist, so expect a no-op until tests are added.

## Coding Style & Naming Conventions
- Go 1.25 toolchain; run `gofmt -w .` before committing. Prefer `go vet ./...` for quick sanity checks.  
- Exported types/functions use PascalCase; locals and unexported helpers use camelCase. File names follow the package role (`crypto/aes_gcm.go`, `parser/env.go`).  
- Configuration keys and CLI flags mirror existing names (`backend`, `recipients`, `secrets`, `-a`, `-i`, `-o`); keep new flags short and kebab-cased.

## Testing Guidelines
- Place `_test.go` files alongside implementations; standard `testing` package is expected.  
- Favor table-driven tests for encryption/decryption permutations and project workflows (init → add-recipient → encrypt-all).  
- Target high coverage on crypto interface boundaries and `.env` parsing to catch regressions; run `go test ./...` (optionally `-race`) before opening a PR.

## Commit & Pull Request Guidelines
- Git history shows Conventional Commit style (`feat: ...`); stick to `feat|fix|chore|docs|refactor` prefixes with succinct scope, e.g., `fix: handle missing .podx.yaml`.  
- PRs should include: short summary, commands run (`make build`, `go test ./...`), and any platform notes for cross-build changes.  
- Link related issues, and add before/after snippets or sample CLI transcripts when behavior changes (encryption output, init defaults). Avoid committing plaintext secrets; ensure `.env` variants are encrypted (`*.podx`) or gitignored.
