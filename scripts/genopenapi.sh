#!/usr/bin/env bash

set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib.sh"

check_commands oapi-codegen

PROJECT_ROOT=$(get_project_root)
cd "$PROJECT_ROOT"

resolve_openapi_file() {
    local service_name="$1"
    local yml_file="api/openapi/${service_name}-swagger.yml"
    local yaml_file="api/openapi/${service_name}-swagger.yaml"

    if [ -f "$yml_file" ]; then
        printf "%s\n" "$yml_file"
        return 0
    fi
    if [ -f "$yaml_file" ]; then
        printf "%s\n" "$yaml_file"
        return 0
    fi

    return 1
}

clean_openapi_generated() {
    local output_dir="$1"
    log_info "Cleaning old generated files in: $output_dir"
    find "$output_dir" -type f -name "openapi_*.gen.go" -delete
}

clean_legacy_openapi_generated() {
    local legacy_dir="$1"
    if [ -d "$legacy_dir" ]; then
        find "$legacy_dir" -maxdepth 1 -type f -name "openapi_*.gen.go" -delete
    fi
}

generate_openapi_client() {
    local service_name="$1"
    local openapi_file="$2"
    local output_dir="$3"
    local package_name="$4"
    local legacy_dir="$5"

    clean_legacy_openapi_generated "$legacy_dir"
    ensure_dir "$output_dir"
    clean_openapi_generated "$output_dir"

    log_info "Generating OpenAPI client code for service: $service_name"
    log_info "  Package: $package_name"
    log_info "  Output: $output_dir"

    oapi-codegen -generate types -o "$output_dir/openapi_types.gen.go" -package "$package_name" "$openapi_file"
    oapi-codegen -generate client -o "$output_dir/openapi_client.gen.go" -package "$package_name" "$openapi_file"
}

generate_openapi_server() {
    local service_name="$1"
    local openapi_file="$2"
    local output_dir="$3"
    local package_name="$4"
    local legacy_dir="$5"

    clean_legacy_openapi_generated "$legacy_dir"
    ensure_dir "$output_dir"
    clean_openapi_generated "$output_dir"

    log_info "Generating OpenAPI server code for service: $service_name"
    log_info "  Package: $package_name"
    log_info "  Output: $output_dir"

    oapi-codegen -generate types -o "$output_dir/openapi_types.gen.go" -package "$package_name" "$openapi_file"
    oapi-codegen -generate gin -o "$output_dir/openapi_api.gen.go" -package "$package_name" "$openapi_file"
}

generate_openapi_for_service() {
    local service_name="$1"
    local openapi_file

    if ! openapi_file="$(resolve_openapi_file "$service_name")"; then
        log_warn "OpenAPI file not found for service: $service_name"
        return 0
    fi

    log_info "⚠️ openapi_file is $openapi_file"

    generate_openapi_client \
        "$service_name" \
        "$openapi_file" \
        "common/client/$service_name" \
        "$service_name" \
        "common/client/$service_name"

    if [ -d "$service_name" ]; then
        generate_openapi_server \
            "$service_name" \
            "$openapi_file" \
            "$service_name/protos" \
            "protos" \
            "$service_name/protos"
    else
        log_warn "Service directory not found: $service_name, skipped server stubs"
    fi
}

log_info "Generating OpenAPI code..."

found_openapi=false
service_names="$(
    find api/openapi -maxdepth 1 -type f \( -name "*-swagger.yml" -o -name "*-swagger.yaml" \) \
        -exec basename {} \; \
        | sed -E 's/-swagger\.ya?ml$//' \
        | sort -u
)"

for service_name in $service_names; do
    found_openapi=true
    log_info "Processing service: $service_name"
    generate_openapi_for_service "$service_name"
done

if [ "${found_openapi}" = false ]; then
    log_warn "No OpenAPI swagger files found in api/openapi"
    exit 0
fi

log_info "✅ OpenAPI code generation completed!"

log_info "Running go mod tidy..."
run_go_mod_tidy "$PROJECT_ROOT"

log_info "✅ go mod tidy completed!"
