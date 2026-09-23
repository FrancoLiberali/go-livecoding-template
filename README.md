# go-interview

Scaffold for a Go live-coding interview. Two independent, self-contained
services (HTTP and gRPC), each in a **controller → service → repository**
layered architecture with interfaces at every seam so each layer is unit-tested
in isolation with [mockery](https://vektra.github.io/mockery/)-generated mocks.

## Layout

```
cmd/
  httpserver/        HTTP entrypoint  (wires repo → service → controller, chi)
  grpcserver/        gRPC entrypoint  (wires repo → service → controller, grpc)
internal/
  httpsvc/           HTTP service (independent tree)
    domain/          entity + sentinel errors
    repository/      Repository interface + in-memory impl   (+ mocks/)
    service/         Service interface + business logic       (+ mocks/)
    controller/      chi HTTP handlers (depend on Service)
  grpcsvc/           gRPC service (independent tree, same shape)
    domain/  repository/  service/  controller/  pb/ (generated)
api/proto/           .proto definitions
.mockery.yaml        mock generation config
.golangci.yml        full strict linter (from the cql project)
```

Each layer depends on the **interface** of the layer below it:
`controller → service.Service → repository.Repository`. Tests mock the
dependency: service tests mock `Repository`, controller tests mock `Service`.

## Validation & structured errors

Request bodies are validated with
[go-playground/validator](https://github.com/go-playground/validator) via struct
tags (e.g. `validate:"required,max=100"`). Field names in messages come from the
`json` tag, so they match the payload.

**Every HTTP response is JSON.** All error responses share one envelope:

```json
{ "error": { "code": "validation_error",
             "message": "request validation failed",
             "details": [ { "field": "name", "message": "name is required" } ] } }
```

| Situation            | Status | `code`             |
|----------------------|--------|--------------------|
| malformed JSON body  | 400    | `invalid_request`  |
| failed validation    | 400    | `validation_error` (with `details`) |
| item not found       | 404    | `not_found`        |
| endpoint timed out   | 504    | `timeout`          |
| unexpected failure   | 500    | `internal`         |

The gRPC service uses the same validator and maps to status codes:
`InvalidArgument` (validation), `NotFound`, `DeadlineExceeded` (timeout),
`Internal`.

## Per-endpoint timeouts

Each endpoint/RPC runs under its own timeout (`controller.Timeouts`, defaults in
`DefaultTimeouts()`). The call runs via `runWithTimeout`, which `select`s on the
deadline vs. the result — so a slow request yields **504** (HTTP) /
**DeadlineExceeded** (gRPC) even if a downstream ignores the context. Use
`controller.NewWithTimeouts(svc, ...)` to inject short timeouts in tests.

## Commands

```
make run-http      # start HTTP server on :8080
make run-grpc      # start gRPC server on :9090 (reflection enabled)
make testv         # all tests, race detector, pretty output (gotestsum)
make lint          # golangci-lint (strict)
make fmt           # apply formatters (gci + gofumpt)
make generate      # regenerate proto stubs + mocks
make proto         # regenerate gRPC stubs only
make mocks         # regenerate mocks only
make tools         # install codegen/lint tooling
```

## Adding a new interface + mock

1. Define the `interface` in its package.
2. Add it under `packages:` in `.mockery.yaml`.
3. `make mocks` → mock lands in that package's `mocks/` subdir.
4. In tests: `m := mocks.NewMockXxx(t); m.EXPECT().Method(args).Return(...)`.

## Quick manual checks

```
# HTTP
curl -s -XPOST localhost:8080/items -d '{"name":"widget"}'
curl -s localhost:8080/items/<id>

# gRPC (reflection)
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext -d '{"name":"widget"}' localhost:9090 grpcsvc.v1.ItemService/CreateItem
```

The repo root (`main.go` / `main_test.go`) is a scratch pad for
algorithm-style problems that don't need the service layout.
