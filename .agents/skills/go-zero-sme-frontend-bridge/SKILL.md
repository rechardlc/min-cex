---
name: go-zero-sme-frontend-bridge
description: Plan, explain, review, and implement Golang go-zero services and project structure for small and medium enterprise backends. Use when Codex works on `go-zero-demo` or similar go-zero projects for architecture, directory design, API/RPC layering, code generation follow-up, refactor, bugfix, tests, documentation, or learning support. Prioritize this skill when the user is frontend-first and wants frontend analogies, learning-oriented comments, progress tracking updates, and solutions that follow practical SME go-zero conventions.
---

# Go Zero Sme Frontend Bridge

Help a frontend-first developer learn and build `go-zero` projects without losing engineering rigor.

Treat every task as both:
- Delivery work: keep the project aligned with practical SME `go-zero` structure
- Teaching work: explain important Go and go-zero concepts with frontend analogies
- Documentation work: update progress tracking whenever new project content is added

## Workflow

1. Detect whether the task touches `go-zero-demo` or another `go-zero` Go service.
2. Read the project planning docs before making structure decisions.
3. Map the request to `go-zero` layers such as API, RPC, `handler`, `logic`, `svc`, `config`, `pkg`, and deployment files.
4. Keep additions aligned with SME backend conventions rather than toy examples.
5. Add short frontend-to-Go comments in changed Go files when the logic is non-obvious.
6. Update progress documentation after every meaningful addition.

## Required behavior

### Always update progress docs

After adding new files, modules, services, or learning artifacts, update project progress documentation in the same turn.

Prefer updating:
- `go-zero-demo/PROGRESS_TRACKER.md`
- `go-zero-demo/docs/LEARNING_PATH.md` when the learning route or stage changes

Do not leave project growth undocumented.

### Explain with frontend analogies

When editing or explaining Go or go-zero code for this user:
- Compare `handler` to a frontend controller or route event entry
- Compare `logic` to a page-level business action or a server action/use-case layer
- Compare `svc` to a dependency container or app context provider
- Compare `config` to environment configuration plus app bootstrap settings
- Compare `pkg` shared utilities to frontend shared libs or `utils/`
- Compare RPC service boundaries to frontend BFF-to-service calls or internal API contracts

For non-obvious changed Go code, add concise comments in this format:

```go
// FE analogy: similar frontend mental model
// Go detail: what is different in Go/go-zero and why this pattern exists
```

Add these comments only where they reduce learning friction.

### Follow SME go-zero conventions

Prefer conventions that scale for small and medium enterprise teams:
- Separate services by responsibility under `apps/`
- Keep `handler` thin and `logic` focused on use-cases
- Use `svc` as the dependency assembly point
- Keep shared concerns in `pkg/` and avoid leaking business logic into it
- Prefer explicit configuration files and typed config structs
- Use `go.work` with per-service `go.mod` when the project is intentionally multi-service
- Keep deployment concerns under `deploy/`
- Keep learning docs and project docs close to the code they explain

### Choose the right delivery mode

For learning-only requests:
- Prioritize explanation, structure, examples, and tradeoffs
- Avoid adding unnecessary production complexity

For implementation requests:
- Deliver runnable or near-runnable structure
- Preserve `go-zero` layering
- Keep placeholder code minimal but intentional

For review requests:
- Prioritize architectural risk, broken layering, dependency leakage, and missing docs updates

## What to read

Read these references from this skill when needed:
- `references/sme-go-zero-rules.md` for project structure and update rules
- `references/frontend-mental-map.md` for frontend-to-go-zero concept mapping

Read these project docs before large structure changes:
- `go-zero-demo/DIRECTORY_RULES.md`
- `go-zero-demo/docs/LEARNING_PATH.md`
- `go-zero-demo/PROGRESS_TRACKER.md`

## Final response checklist

Include:
- What changed
- Which `go-zero` layer or boundary was involved
- At least 2 frontend analogies when Go or go-zero concepts matter
- Whether progress docs were updated
- Any remaining learning or implementation next step
