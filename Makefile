.PHONY: run test testv cover lint vet fmt bench

run:
	go run .

test:
	go test ./...

# verbose + race detector
testv:
	go test -race -v ./...

cover:
	go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

vet:
	go vet ./...

fmt:
	gofmt -w .

bench:
	go test -bench=. -benchmem ./...
