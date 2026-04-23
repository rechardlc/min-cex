---
name: go-server-frontend-bridge
description: Bridge frontend mental models to Go backend implementation for beginner Go developers. Use this skill whenever the user asks for Golang work in the `server/` directory, including coding, refactor, bugfix, review, architecture, tests, or explanations. Prioritize this skill when the user has frontend experience and asks for side-by-side frontend vs Go understanding. This skill must produce comparative inline comments in changed Go code and comparative explanations in the final response.
---

# Go Server Frontend Bridge

Help a frontend-first developer learn Go while shipping production code in `server/`.

## Core behavior

Treat every Go task in `server/` as both:
- Delivery work (make the code correct and runnable)
- Teaching work (map each important Go idea to a frontend analogy)

If the request does not touch Go files in `server/`, do not force this skill.

## Workflow

1. Detect scope
- Confirm the change set includes `server/` and Go-related files (`.go`, `go.mod`, `go.sum`, or Go tests).
- Continue with normal engineering flow if scope matches.

2. Plan with concept mapping
- Before editing, identify the key concepts involved (for example: concurrency, context, interfaces, error handling, pointer/value semantics, package boundaries).
- Use matching analogy patterns in `references/frontend-go-map.md`.

3. Implement with comparative inline comments
- Add short comparative comments for non-obvious logic in changed Go code.
- Use this comment pair format near the relevant block:

```go
// FE analogy: <similar frontend or JS/TS concept>
// Go detail: <what is different in Go and why this pattern is used>
```

- Keep comments concise and practical.
- Do not add noisy comments for obvious one-liners.

4. Explain changes with a frontend bridge summary
- In the final response, include:
  - What changed (files and behavior)
  - Frontend-to-Go concept mapping for this task
  - Why the Go approach is idiomatic
  - How to run and verify

## Required final response checklist

For Go changes in `server/`, include all items:

1. Changed files and purpose
2. At least 3 frontend vs Go mappings (or all key mappings if fewer than 3 exist)
3. One "watch out" section for common beginner mistakes
4. Verification commands and expected outcome

## Comment placement guidance

Always add comparative comments when introducing or changing:
- Goroutines, channels, or worker pools
- `context.Context` propagation or cancellation
- Interface-based design or dependency injection patterns
- Error wrapping/propagation (`fmt.Errorf`, sentinel errors, typed errors)
- Pointer vs value behavior that affects correctness or performance
- Request lifecycle logic in handlers or middleware

Prefer short explanations in code and deeper explanation in the final response.

## Quality bar

- Keep business logic clean; comments should teach, not clutter.
- Use idiomatic Go even when drawing frontend analogies.
- Never sacrifice correctness for analogy simplicity.

## Reference

Read `references/frontend-go-map.md` when selecting analogies for explanations.
