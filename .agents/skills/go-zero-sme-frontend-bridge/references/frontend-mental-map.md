# Frontend mental map for go-zero

## Core analogies

- `handler` ~= route handler / controller entry in a frontend BFF
- `logic` ~= page action, server action, or use-case function
- `svc.ServiceContext` ~= app-level dependency container or provider registry
- `config` ~= environment variables plus app bootstrap config
- `pkg/response` ~= shared response helper used across pages or APIs
- `pkg/xerr` ~= centralized frontend error enum or shared error helper
- RPC service ~= internal service contract, similar to typed calls between frontend BFF and backend domains

## Important differences

- Go uses packages and explicit types where frontend code may rely more on convention and flexible objects
- `go-zero` encourages stronger layering than many ad-hoc Node services
- Dependency injection is usually done through structs rather than runtime container magic
- Error handling is explicit and returned, not thrown by default

## Comment pattern

Use short comments only for the parts that are likely to confuse a frontend-first learner:

```go
// FE analogy: this is like ...
// Go detail: here we use ...
```
