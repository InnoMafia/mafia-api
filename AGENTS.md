# Working on mafia-api

## Scope

- This repository owns the Go API and game rules for one self-hosted club.
- Keep it a modular monolith. Do not introduce microservices, a message broker,
  Kubernetes, multi-tenancy, or a database migration without a concrete requirement.
- Judge UI, OBS rendering, and browser-extension code belong in `mafia-web`,
  `mafia-hud`, and `mafia-extension`, respectively.
- `InnoMafia/mafia-old` is the behavioral reference, not a dependency to run in
  production alongside the new game engine.

## Layout and design

- `cmd/mafia-api` wires configuration, signals, logging, and the server.
- `internal/config` parses environment variables without changing process state.
- `internal/httpapi` owns HTTP adapters; `internal/httpserver` owns their lifecycle.
- Add game-domain and persistence packages only when implementing real behavior.
  Do not scaffold unused layers, interfaces, or generated implementations.
- Prefer the standard library and small, explicit dependencies. Use `context`
  for cancellation and `log/slog` for structured logging.
- Keep game rules on the server. Clients receive role-specific projections.
  Never expose the full judge state through a public or HUD endpoint by accident.
- When contracts are introduced, keep their source here, version changes, and
  generate a TypeScript SDK for the other repositories. Do not duplicate schemas
  manually or make consumers depend on an unversioned branch.

## Checks and tests

- Use the Go version in `go.mod`; keep the Docker builder and CI consistent.
- Run `make fmt` and `make check` after changes. Build and smoke-test the image
  when changing container or startup behavior.
- Put tests beside their implementation in `foo_test.go`.
- Use readable table-driven tests, explicit inputs and expectations, and
  `t.Parallel()` with case-local mutable state. Do not mutate process environment
  in parallel tests; inject configuration lookup instead.
- Check observable results and error classifications, not error-message strings.
  Use real HTTP servers for lifecycle behavior and real storage for persistence.
- Do not add dependencies solely for mocks or assertions when standard-library
  tests are sufficient. Never weaken assertions to make a failing check pass.

## Security and operation

- The current scaffold has no authentication or game endpoints.
- Bind locally by default. Containers listen on all interfaces internally, while
  Compose publishes only to loopback by default.
- Do not commit secrets, local environment files, production databases, player
  photos, or real exported protocols. Use synthetic fixtures.
- Judge and HUD links are access credentials. Do not log their tokens, cookies,
  passwords, or complete sensitive request payloads.
- Health checks must not reveal secrets. Readiness currently checks only the
  HTTP process; add dependency checks when those dependencies actually exist.
- Review HTTP deadlines when adding WebSocket or streaming endpoints; preserve
  long-lived connections and implement their shutdown explicitly.

## Releases and changes

- `deploy/` owns the club's deployment composition. Each application builds its
  own image; a release pins a tested set of image versions or digests.
- Do not deploy by pulling several default branches or using mutable `latest`
  tags. The browser extension has its own release cycle.
- Do not push, publish packages/images, change GitHub settings, or deploy unless
  the user authorizes that operation. CI currently checks and builds only.
- Review the final diff and run relevant checks. Create a local commit for
  completed authorized changes unless the user asks otherwise; stage only task
  changes and report any checks that could not run.
- During the future legacy migration, preserve game IDs, links, and data.
  Never let two engines write the same data directory. Test rollback after new
  writes, rather than assuming an old backup preserves subsequent actions.
