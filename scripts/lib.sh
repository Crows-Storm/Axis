#!/usr/bin/env bash

set -u

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    printf "%b[INFO]%b %s\n" "${GREEN}" "${NC}" "$1"
}

log_error() {
    printf "%b[ERROR]%b %s\n" "${RED}" "${NC}" "$1" >&2
}

log_warn() {
    printf "%b[WARN]%b %s\n" "${YELLOW}" "${NC}" "$1"
}

die() {
    log_error "$1"
    exit 1
}

check_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        die "$1 is not installed, please install it first"
    fi
}

check_commands() {
    local cmd
    for cmd in "$@"; do
        check_command "$cmd"
    done
}

get_project_root() {
    local caller_dir
    caller_dir="$(cd "$(dirname "${BASH_SOURCE[1]}")" && pwd)"
    (cd "${caller_dir}/.." && pwd)
}

ensure_dir() {
    local dir_path="$1"
    if [ ! -d "$dir_path" ]; then
        mkdir -p "$dir_path"
        log_info "Created directory: $dir_path"
    fi
}

run_go_mod_tidy() {
    local project_root="$1"
    local module_dir
    local cache_dir=""
    local has_module=false

    if ! command -v go >/dev/null 2>&1; then
        log_warn "go is not installed, skipping go mod tidy"
        return 0
    fi

    if [ -n "${GOCACHE:-}" ]; then
        cache_dir="${GOCACHE}"
    else
        cache_dir="$(go env GOCACHE 2>/dev/null || true)"
    fi

    if [ -z "${cache_dir}" ] || [ ! -d "${cache_dir}" ] || [ ! -w "${cache_dir}" ]; then
        cache_dir="${project_root}/.cache/go-build"
        ensure_dir "${cache_dir}"
        log_info "Using local go build cache: ${cache_dir}"
    fi

    if [ "${GOCACHE:-}" != "${cache_dir}" ]; then
        export GOCACHE="${cache_dir}"
    fi

    if [ -f "${project_root}/pkg/go.mod" ]; then
        has_module=true
        log_info "Tidying module: pkg"
        (cd "${project_root}/pkg" && go mod tidy)
    fi

    for module_dir in "${project_root}"/*; do
        [ -d "${module_dir}" ] || continue
        [ -f "${module_dir}/go.mod" ] || continue
        has_module=true
        log_info "Tidying module: $(basename "${module_dir}")"
        (cd "${module_dir}" && go mod tidy)
    done

    if [ "${has_module}" = false ]; then
        log_warn "No go.mod files found, skipped go mod tidy"
    fi
}
