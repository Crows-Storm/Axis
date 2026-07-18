#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib.sh"

check_commands protoc protoc-gen-go protoc-gen-go-grpc

PROJECT_ROOT=$(get_project_root)
cd "$PROJECT_ROOT"

gen_proto_for_output() {
    local service_name="$1"
    local output_dir="$2"

    ensure_dir "$output_dir"

    log_info "Cleaning old generated files in: $output_dir"
    find "$output_dir" -type f -name "*.pb.go" -delete

    log_info "Generating protobuf code for service: $service_name"
    log_info "  Output: $output_dir"

    # Find the proto directory
    local proto_dir=""
    for dir in api/*pb; do
        if [ -f "$dir/${service_name}.proto" ]; then
            proto_dir="$dir"
            break
        fi
    done

    protoc \
        --proto_path="$proto_dir" \
        --go_out="$output_dir" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$output_dir" \
        --go-grpc_opt=paths=source_relative \
        "${service_name}.proto"
}

generate_proto_for_service() {
    local service_name="$1"
    local proto_file=""

    # Find the proto file in api/*pb directories
    for dir in api/*pb; do
        if [ -f "$dir/${service_name}.proto" ]; then
            proto_file="$dir/${service_name}.proto"
            break
        fi
    done

    if [ -z "$proto_file" ] || [ ! -f "$proto_file" ]; then
        log_warn "Proto file not found for service: $service_name"
        return 0
    fi

    # Generate to common/genproto/{serviceName}pb
    gen_proto_for_output "$service_name" "common/genproto/${service_name}pb"
}

log_info "Generating protobuf code..."

found_proto=false
for proto_file in api/*pb/*.proto; do
    [ -f "$proto_file" ] || continue
    found_proto=true
    service_name="$(basename "$proto_file" .proto)"
    log_info "Processing service: $service_name"
    generate_proto_for_service "$service_name"
done

if [ "${found_proto}" = false ]; then
    log_warn "No proto files found in api/*pb"
    exit 0
fi

log_info "✅ Protobuf code generation completed!"

log_info "Running go mod tidy..."
run_go_mod_tidy "$PROJECT_ROOT"

log_info "✅ go mod tidy completed!"
