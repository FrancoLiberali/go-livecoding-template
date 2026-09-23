.PHONY: run test testv cover lint lintfix vet fmt bench

run:
	go run .

test:
	go test ./...

# pretty per-test output via gotestsum, with race detector
testv:
	gotestsum --format testname -- -race ./...

cover:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

lintfix:
	golangci-lint run --fix

vet:
	go vet ./...

fmt:
	gofmt -w .

bench:
	go test -bench=. -benchmem ./...
