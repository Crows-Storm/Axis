.PHONY: gen genproto genopenapi gotidy clean help \
        test test-auth test-auth-coverage test-auth-grpc test-bench \
        gen-mock gen-mock-svc clean-mock clean-mock-svc

.DEFAULT_GOAL := help

MOCK_SCRIPT := scripts/test_mock.sh

gen: genproto genopenapi gen-mock gotidy
	@echo "✅ All code generation completed!"

genproto:
	@echo "🔨 Generating protobuf code..."
	@bash scripts/genproto.sh

genopenapi:
	@echo "🔨 Generating OpenAPI code..."
	@bash scripts/genopenapi.sh

gen-mock:
	@bash $(MOCK_SCRIPT) gen

gen-mock-svc:
	@if [ -z "$(SVC)" ]; then echo "❌ Usage: make gen-mock-svc SVC=<service>"; exit 1; fi
	@bash $(MOCK_SCRIPT) gen $(SVC)

# ==================== Clean ====================
clean: clean-mock
	@echo "🧹 Cleaning generated code..."
	@find common/client -name "*.gen.go" -type f -delete 2>/dev/null || true
	@find ./ -name "*.gen.go" -type f -delete 2>/dev/null || true
	@find ./ -name "*.pb.go" -type f -delete 2>/dev/null || true
	@echo "✅ Cleanup completed!"

clean-mock:
	@bash $(MOCK_SCRIPT) clean

clean-mock-svc:
	@if [ -z "$(SVC)" ]; then echo "❌ Usage: make clean-mock-svc SVC=<service>"; exit 1; fi
	@bash $(MOCK_SCRIPT) clean $(SVC)

# ==================== Testing ====================
test:
	@echo "🧪 Running all tests..."
	@for dir in $$(find . -name "go.mod" -not -path "*/dagger/*" -exec dirname {} \;); do \
		echo "  → Testing $$dir"; \
		(cd "$$dir" && go test -v -race ./...) || exit 1; \
	done
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

# ==================== Maintenance ====================
gotidy:
	@find ./ -name "go.mod" -not -path "*/dagger/*" -exec dirname {} \; | while read dir; do \
		echo "Running go mod tidy in $$dir..."; \
		(cd "$$dir" && go mod tidy) || exit 1; \
	done
	@echo "✅ All go modules tidied!"

help:
	@echo "============================================"
	@echo "  Axis Microservices - Available Commands"
	@echo "============================================"
	@echo ""
	@echo "📦 Code Generation:"
	@echo "  make gen                  - Generate all code (protobuf + OpenAPI + mocks)"
	@echo "  make genproto             - Generate protobuf code only"
	@echo "  make genopenapi           - Generate OpenAPI code only"
	@echo "  make gen-mock             - Generate mocks for ALL services"
	@echo "  make gen-mock-svc SVC=x   - Generate mocks for specific service"
	@echo ""
	@echo "🧹 Clean:"
	@echo "  make clean                - Clean all generated code (includes mocks)"
	@echo "  make clean-mock           - Clean mocks for ALL services"
	@echo "  make clean-mock-svc SVC=x - Clean mocks for specific service"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  make test                 - Run all tests across all modules"
	@echo "  make test-auth            - Run auth service tests"
	@echo "  make test-auth-coverage   - Run auth tests with coverage report"
	@echo "  make test-auth-grpc       - Run auth gRPC tests only"
	@echo "  make test-bench           - Run benchmark tests"
	@echo ""
	@echo "🔧 Maintenance:"
	@echo "  make gotidy               - Run go mod tidy on all modules"
	@echo "  make help                 - Show this help message"
	@echo ""