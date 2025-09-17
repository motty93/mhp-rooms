# Repository Guidelines

## Project Structure & Module Organization
Application entrypoints live in `cmd/` (`cmd/server` for the web server, `cmd/migrate` plus `cmd/seed` for database tasks). Core domain packages sit under `internal/`, with `internal/handlers` for HTTP flow, `internal/models` for GORM entities, and `internal/repository` for persistence logic. Templated views are stored in `templates/` (`layouts/`, `pages/`, `components/`), while static assets ship from `static/`. Auxiliary docs live in `docs/`, and build artifacts land in `bin/`. Keep new assets and partials aligned with this layout to simplify discovery.

## Build, Test, and Development Commands
Run `make setup` when onboarding; it installs dependencies, starts Docker services, migrates schemas, and seeds data. Day-to-day, `make dev` starts the server with Air hot reload, whereas `make run` performs a one-off build and execution. Database workflows use `make migrate` for compiled migrations and `make seeds` for fixtures. For container troubleshooting, `docker compose up -d` restarts Postgres and `make container-reset` rebuilds the stack.

## Coding Style & Naming Conventions
Go code must stay `gofmt`-clean with mixedCaps identifiers and GoDoc comments on exported APIs. Shared helpers belong in `internal/utils` to avoid package cycles. Front-end snippets follow the Biome config (`space` indentation, single quotes, semicolons only when required, 110-character lines). Template partials use snake_case filenames and group Tailwind utility classes logically to retain readability.

## Testing Guidelines
Author Go tests alongside implementations using the `_test.go` suffix and table-driven cases. Before pushing, run `go test -v ./...` and confirm coverage-sensitive paths with `go test -cover ./...`. Prefer fake repositories for unit isolation, but spin up the Docker database (`make container-up`) for integration flows involving GORM or SQL queries. Validate session handling, room filters, and template rendering edge cases whenever handlers change.

## Commit & Pull Request Guidelines
Commits use Conventional Commit prefixes (`feat:`, `fix:`, `refactor:`, etc.) and the imperative mood. Keep related changes grouped and describe context in Japanese, matching the rest of the history. Pull requests should summarize the change, list manual verification steps (for example `make test`, browser smoke checks), reference linked issues, and attach UI screenshots or recordings when templates or CSS shift. Call out schema updates, migrations, or seed adjustments explicitly so reviewers can apply them locally.

## Collaboration Notes
Team communication, inline comments, and documentation default to Japanese. Confirm any adjustments to header layout or other UX-critical conventions with product stakeholders before merging.
