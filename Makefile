.PHONY: run test check clean

# Build server and start it
run:
	@go build -o bin/tavrn-admin ./cmd/tavrn-admin
	@./bin/tavrn-admin & sleep 1 && ssh localhost -p 2222; kill %1 2>/dev/null

# Run all tests with race detector
test:
	go test -race ./internal/... ./ui/...

# Run before push — lint + build + test (mirrors CI)
check:
	gofmt -w .
	go vet ./...
	go build -o bin/tavrn-admin ./cmd/tavrn-admin
	go test -race ./internal/... ./ui/...
	@echo "All good."

# Remove binaries and db
clean:
	rm -rf bin/ tavrn.db
