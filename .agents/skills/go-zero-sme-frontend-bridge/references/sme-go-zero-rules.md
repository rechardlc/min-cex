# SME go-zero rules

## Scope

Use these rules when planning, explaining, or implementing `go-zero` work in this project.

## Project structure

- Keep service code under `apps/`
- Keep API-facing services and RPC services separated
- Keep reusable code under `pkg/`
- Keep deployment and environment files under `deploy/`
- Keep learning and planning materials in project docs

## Layering

- `cmd/`: service entrypoint
- `api/` or `desc/`: contract definitions
- `internal/config/`: typed service config
- `internal/handler/`: request binding and response return
- `internal/logic/`: business use-case implementation
- `internal/svc/`: dependency wiring

## Coding rules

- Keep `handler` thin
- Keep `logic` focused on one use-case
- Put shared helpers in `pkg/` only when they are truly cross-service
- Prefer explicit naming over clever naming
- Keep placeholder code small and educational

## Documentation update rule

Whenever new files, folders, services, or major learning artifacts are added:
- update `go-zero-demo/PROGRESS_TRACKER.md`
- update `go-zero-demo/docs/LEARNING_PATH.md` if the learning sequence changes

## Learning rule

Prefer code and comments that help a frontend-first developer build accurate mental models of:
- API vs RPC
- `handler` vs `logic`
- dependency injection through `svc`
- project-level module boundaries
