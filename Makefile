.PHONY: test lint fmt vet integration clean

# Run unit tests (short mode, no integration)
test:
	go test -v -short ./...

# Run tests with coverage
test-cov:
	go test -v -short -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Run integration tests (requires CUBRID)
integration:
	go test -v -run Integration ./...

# Run go vet
vet:
	go vet ./...

# Run golangci-lint (if installed)
lint: vet
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

# Format code
fmt:
	gofmt -w .

# Clean build artifacts
clean:
	rm -f coverage.out

# Full check
check: fmt vet test
