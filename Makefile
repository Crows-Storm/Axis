.PHONY: gen genproto genopenapi gotidy clean help test test-auth test-auth-coverage test-auth-grpc test-bench

.DEFAULT_GOAL := help

gen: genproto genopenapi
	@echo "✅ All code generation completed!"

genproto:
	@echo "🔨 Generating protobuf code..."
	@bash scripts/genproto.sh

genopenapi:
	@echo "🔨 Generating OpenAPI code..."
	@bash scripts/genopenapi.sh

clean:
	@echo "🧹 Cleaning generated code..."
	@find common/client -name "*.gen.go" -type f -delete 2>/dev/null || true
	@find ./ -name "*.gen.go" -type f -delete 2>/dev/null || true
	@find ./ -name "*.pb.go" -type f -delete 2>/dev/null || true
	@echo "✅ Cleanup completed!"

gotidy:
	@find ./ -name "go.mod" -type f -exec dirname {} \; | while read dir; do \
		echo "Running go mod tidy in $$dir..."; \
		cd "$$dir" && go mod tidy && cd - > /dev/null; \
	done
	@echo "All go modules tidied!"

test:
	@echo "🧪 Running all tests..."
	@go test -v -race ./...
	@echo "✅ All tests completed!"

test-auth:
	@echo "🧪 Running auth service tests..."
	@cd auth && go test -v -race ./...
	@echo "✅ Auth tests completed!"

test-auth-coverage:
	@echo "📊 Generating coverage report for auth service..."
	@cd auth && go test -v -race -coverprofile=coverage.out ./...
	@cd auth && go tool cover -func=coverage.out
	@cd auth && go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: auth/coverage.html"

test-auth-grpc:
	@echo "🧪 Running auth gRPC tests..."
	@cd auth && go test -v -race ./adapters/grpc/...
	@echo "✅ Auth gRPC tests completed!"

test-bench:
	@echo "⚡ Running benchmark tests..."
	@cd auth && go test -bench=. -benchmem ./...
	@echo "✅ Benchmark tests completed!"

help:
	@echo "Available commands:"
	@echo "  make gen                  - Generate all code (protobuf + OpenAPI)"
	@echo "  make genproto             - Generate protobuf code only"
	@echo "  make genopenapi           - Generate OpenAPI code only"
	@echo "  make clean                - Clean all generated code"
	@echo "  make gotidy               - Run go mod tidy on all modules"
	@echo "  make test                 - Run all tests"
	@echo "  make test-auth            - Run auth service tests"
	@echo "  make test-auth-coverage   - Run auth tests with coverage report"
	@echo "  make test-auth-grpc       - Run auth gRPC tests only"
	@echo "  make test-bench           - Run benchmark tests"
	@echo "  make help                 - Show this help message"