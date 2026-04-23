# Frontend to Go Concept Map

Use this sheet to pick analogies quickly. Reword based on the current task.

## Core mappings

| Frontend/JS concept | Go concept | Explain with this angle |
| --- | --- | --- |
| Event loop + async task queue | Goroutines + scheduler | Go runs many lightweight tasks concurrently without manual promise chaining. |
| `Promise.all` for fan-out work | `sync.WaitGroup` + goroutines | Start many units of work, then wait for all to finish. |
| `AbortController` | `context.Context` cancellation | Propagate cancellation and timeout through call chains. |
| Prop drilling request state | Context passed by parameter | Make request-scoped data explicit and cancellable. |
| TypeScript interfaces | Go interfaces | Go interfaces are satisfied implicitly; no `implements` keyword. |
| Class instance with methods | Struct with receiver methods | Data + behavior without inheritance-heavy hierarchy. |
| React composition | Struct embedding/composition | Prefer composition over inheritance in both ecosystems. |
| `try/catch` exceptions | Explicit `error` returns | Handle failures where they occur; keep control flow visible. |
| Throw wrapped error objects | `fmt.Errorf("...: %w", err)` | Preserve root cause while adding context. |
| Mutable object reference behavior | Pointer vs value semantics | Choose pointer receiver when shared mutation or copy cost matters. |
| NPM package boundaries | Go packages/modules | Keep API surface small and package ownership clear. |
| Middleware chain in Express | HTTP middleware in Go | Layer cross-cutting concerns around handlers. |
| Dependency injection containers | Interface + constructor injection | Wire dependencies explicitly for testability. |
| Jest table tests | Go table-driven tests | Express many scenarios as data rows in one test pattern. |
| ESLint/TS compile checks | `go vet` + compiler checks | Lean on built-in tooling and strict compile-time guarantees. |

## Comment templates

Use one of these forms near non-obvious logic:

```go
// FE analogy: Similar to Promise.all for parallel async calls.
// Go detail: Use WaitGroup to block until all goroutines complete.
```

```go
// FE analogy: Like AbortController passed through async layers.
// Go detail: Carry context.Context so timeout/cancel reaches DB and downstream calls.
```

```go
// FE analogy: Similar to TS interface-based dependency mocking.
// Go detail: Depend on interface contracts to keep handlers testable.
```

## Beginner pitfalls to highlight

- Launching goroutines without cancellation strategy.
- Ignoring returned errors or collapsing error context.
- Copying large structs by value unintentionally.
- Storing request-scoped data in globals instead of function parameters/context.
- Overengineering class-like hierarchies instead of small interfaces and composition.
