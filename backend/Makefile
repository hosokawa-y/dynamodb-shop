.PHONY: fmt run build test

# Go format
fmt:
	@echo "Formatting Go files..."
	@gofmt -w .
	@echo "Done."

# Run the application
run:
	@go run cmd/api/main.go

# Build the application
build:
	@go build -o bin/api cmd/api/main.go

# Run tests
test:
	@go test ./...
