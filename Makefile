.PHONY: run-http run-grpc test testv cover lint lintfix fmt vet bench proto mocks generate tools

# --- run the services ---
run-http:
	go run ./cmd/httpserver

run-grpc:
	go run ./cmd/grpcserver

# plain executable (no HTTP/gRPC) for algorithm-style problems
run-cli:
	go run ./cmd/cli

# --- tests ---
test:
	go test ./...

# pretty per-test output via gotestsum, with race detector
testv:
	gotestsum --format testname -- -race ./...

cover:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

# --- lint & format ---
lint:
	golangci-lint run

lintfix:
	golangci-lint run --fix

# apply the configured formatters (gci + gofumpt)
fmt:
	golangci-lint fmt

vet:
	go vet ./...

bench:
	go test -bench=. -benchmem ./...

# --- code generation ---
# regenerate gRPC stubs from api/proto/*.proto into internal/grpcsvc/pb
proto:
	protoc --go_out=. --go_opt=module=interview \
		--go-grpc_out=. --go-grpc_opt=module=interview \
		api/proto/grpcsvc.proto

# regenerate testify mocks from .mockery.yaml
mocks:
	mockery

generate: proto mocks

# install the codegen/lint tooling this repo expects
tools:
	go install github.com/vektra/mockery/v2@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install gotest.tools/gotestsum@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
