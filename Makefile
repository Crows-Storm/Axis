.PHONY: gen genproto genopenapi gotidy clean help

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

help:
	@echo "Available commands:"
	@echo "  make gen         - Generate all code (protobuf + OpenAPI)"
	@echo "  make genproto    - Generate protobuf code only"
	@echo "  make genopenapi  - Generate OpenAPI code only"
	@echo "  make clean       - Clean all generated code"
	@echo "  make help        - Show this help message"