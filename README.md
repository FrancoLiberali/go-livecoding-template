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
    repository/      concrete in-memory store (no interface here)
    service/         business logic + the Repository interface it needs (+ mocks/)
    controller/      chi handlers + the Service interface it needs      (+ mocks/)
  grpcsvc/           gRPC service (independent tree, same shape)
    domain/  repository/  service/  controller/  pb/ (generated)
api/proto/           .proto definitions
.mockery.yaml        mock generation config
.golangci.yml        full strict linter (from the cql project)
```

**Interfaces are declared by their consumer, not their implementer** (Go's
"accept interfaces, return structs" / define-interfaces-where-used idiom):

- `controller` declares the `Service` interface it needs; `service` returns a
  concrete `*Service` that satisfies it.
- `service` declares the `Repository` interface it needs; `repository` returns a
  concrete `*InMemory` that satisfies it.

So the concrete packages depend on nothing above them, and only `cmd/` wires the
chain (`repository → service → controller`). Mocks are generated into each
**consumer** package's `mocks/`: service tests mock `Repository`, controller
tests mock `Service`.

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
| client canceled      | 499    | `canceled`         |
| unexpected failure   | 500    | `internal`         |

The gRPC service uses the same validator and maps to status codes:
`InvalidArgument` (validation), `NotFound`, `DeadlineExceeded` (timeout),
`Canceled` (client cancellation), `Internal`.

**Error mapping is per-endpoint, not centralized.** Each handler decides how its
own domain errors map to a response (the same domain error may warrant a
different status in a different endpoint). Only non-domain failures are shared:
`writeTransportError` / `transportError` handle the timeout (the sole exception)
and the catch-all internal error.

## Resilience & ops

- **gRPC panic recovery** — `grpcmw.UnaryRecovery` recovers handler panics into
  `codes.Internal` (logged with a stack) so a panic can't crash the server
  (grpc-go doesn't recover by default). HTTP is covered by chi's `Recoverer`.
- **Health / readiness** — HTTP `GET /healthz` and `GET /readyz` return
  `{"status":"ok"}`; gRPC exposes the standard `grpc.health.v1.Health` service.
  ```
  curl -s localhost:8080/healthz
  grpcurl -plaintext -d '{"service":"grpcsvc.v1.ItemService"}' localhost:9090 grpc.health.v1.Health/Check
  ```

## Logging

Structured `slog` throughout. Every request/response is logged once:

- **HTTP** — `httpmw.RequestLogger` (a chi middleware) logs `method`, `path`,
  `status`, `bytes`, `duration`, `remote`, `request_id`.
- **gRPC** — `grpcmw.UnaryLogger` (a unary interceptor) logs `method`, `code`,
  `duration`.

> chi ships `middleware.Logger`, but it writes unstructured stdlib-`log` output;
> these thin wrappers keep request logs in the same structured `slog` stream as
> the rest of the app (reusing chi's `WrapResponseWriter`/`GetReqID`).

Unexpected **internal** errors are logged in full (`slog.ErrorContext`) at the
mapping boundary *before* the sanitized `500` / `Internal` is returned — the
client sees a generic message while the real cause survives in the logs.

## Per-endpoint timeouts

Each endpoint/RPC has its own hardcoded timeout (`getItemTimeout` /
`createItemTimeout` constants):

- **HTTP** — a per-route `middleware.Timeout(d)` (chi) sets the deadline on the
  request context.
- **gRPC** — each RPC derives one with `context.WithTimeout(ctx, d)`.

The deadline propagates through `ctx` into the service/repository. When a
ctx-aware downstream honors it, the handler maps the `ctx` error:
`context.DeadlineExceeded` → **504** / **DeadlineExceeded**, and
`context.Canceled` (the client hung up first) → **499** / **Canceled**. The
mapping is covered by fast tests that inject the `ctx` error directly, rather
than sleeping out the real timeout.

> Note: the in-memory repo returns instantly and does **not** honor `ctx`, so a
> real timeout only fires once you back the repo with a ctx-aware datastore (a
> DB driver cancels the query on deadline). The wiring and 504 mapping are in
> place and tested; swapping the repo is all that's needed to make it bite.

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

1. Declare the `interface` in the **consumer** package (the one that calls it).
2. Add that package under `packages:` in `.mockery.yaml`.
3. `make mocks` → mock lands in the consumer package's `mocks/` subdir.
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
