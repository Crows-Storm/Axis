#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

check_mockery() {
    if ! command -v mockery &> /dev/null; then
        echo -e "${RED}❌ mockery not found.${NC}"
        echo -e "${YELLOW}💡 Install: go install github.com/vektra/mockery/v2@latest${NC}"
        exit 1
    fi
}

get_services() {
    local services=()
    for dir in "$ROOT_DIR"/*/; do
        if [ -f "${dir}go.mod" ] && [ -f "${dir}.mockery.yaml" ]; then
            services+=("$(basename "$dir")")
        fi
    done
    echo "${services[@]}"
}

gen_service_mocks() {
    local svc="$1"
    local svc_dir="$ROOT_DIR/$svc"

    if [ ! -f "$svc_dir/.mockery.yaml" ]; then
        echo -e "   ${YELLOW}⚠️ No .mockery.yaml in [$svc], skipping${NC}"
        return
    fi

    echo -e "\n${YELLOW}📦 Processing [$svc]...${NC}"
    if ! (cd "$svc_dir" && mockery 2>&1 | sed 's/^/   /'); then
        echo -e "   ${RED}❌ [$svc] mockery failed!${NC}"
        exit 1
    fi
    echo -e "   ${GREEN}✅ [$svc] Done${NC}"
}

gen_common_mocks() {
    local common_dir="$ROOT_DIR/common"

    if [ ! -f "$common_dir/.mockery.yaml" ]; then
        echo -e "${YELLOW}⚠️ No .mockery.yaml in [common], skipping${NC}"
        return
    fi

    echo -e "\n${YELLOW}📦 Processing [common]...${NC}"
    (
        cd "$common_dir"
        mockery 2>&1 | sed 's/^/   /'
    )
    echo -e "   ${GREEN}✅ [common] Done${NC}"
}

gen_mocks() {
    check_mockery
    local target_service="$1"

    if [ -n "$target_service" ]; then
        if [ ! -d "$ROOT_DIR/$target_service" ]; then
            echo -e "${RED}❌ Service '$target_service' not found.${NC}"
            exit 1
        fi
        gen_service_mocks "$target_service"
    else
        read -ra services <<< "$(get_services)"
        if [ ${#services[@]} -eq 0 ]; then
            echo -e "${RED}❌ No microservices found.${NC}"
            exit 1
        fi

        echo -e "${GREEN}🚀 Generating mocks for: ${CYAN}${services[*]} + common${NC}"

        # 先生成 common（其他服务可能依赖）
        gen_common_mocks

        for svc in "${services[@]}"; do
            gen_service_mocks "$svc"
        done
    fi

    echo -e "\n${GREEN}🎉 All mocks generated successfully!${NC}"
}

clean_mocks() {
    local target_service="$1"
    local total_cleaned=0

    clean_mock_files() {
        local base_dir="$1"
        if [ ! -d "$base_dir" ]; then
            echo "0"
            return
        fi

        local count
        count=$(find "$base_dir" -name "mock_*.go" -type f 2>/dev/null | wc -l | tr -d '[:space:]')

        if [ "$count" -gt 0 ]; then
            find "$base_dir" -name "mock_*.go" -type f -delete
        fi

        # 只输出纯净的数字供外层捕获
        echo "$count"
    }

    if [ -n "$target_service" ]; then
        total_cleaned=$(clean_mock_files "$ROOT_DIR/$target_service")
        if [ "$total_cleaned" -gt 0 ]; then
            echo -e "${GREEN}✅ Cleaned $total_cleaned mock files in $ROOT_DIR/$target_service${NC}"
        fi
    else
        read -ra services <<< "$(get_services)"
        for svc in "${services[@]}"; do
            c=$(clean_mock_files "$ROOT_DIR/$svc")
            if [ "$c" -gt 0 ]; then
                echo -e "${GREEN}✅ Cleaned $c mock files in $ROOT_DIR/$svc${NC}"
            fi
            total_cleaned=$((total_cleaned + c))
        done

        c=$(clean_mock_files "$ROOT_DIR/common")
        if [ "$c" -gt 0 ]; then
            echo -e "${GREEN}✅ Cleaned $c mock files in $ROOT_DIR/common${NC}"
        fi
        total_cleaned=$((total_cleaned + c))
    fi

    if [ "$total_cleaned" -eq 0 ]; then
        echo -e "${YELLOW}⚠️ No mock files found to clean.${NC}"
    else
        echo -e "${GREEN}🧹 Clean complete. Removed $total_cleaned files.${NC}"
    fi
}

show_help() {
    echo -e "${CYAN}TDD Mock Generator (Colocated + YAML Config)${NC}"
    echo ""
    echo "Usage: $0 {gen|clean} [service-name]"
    echo ""
    echo "Each service must have a .mockery.yaml config file."
    echo "Mocks are generated next to interfaces: mock_<snake_case>.go"
    echo ""
    echo "Commands:"
    echo "  gen [svc]   Generate mocks (all or specific service)"
    echo "  clean [svc] Remove mock_*.go files (safe, won't delete source)"
    echo ""
    echo "Examples:"
    echo "  $0 gen          # ALL services + common"
    echo "  $0 gen auth     # Only auth service"
    echo "  $0 clean        # Clean ALL mocks"
    echo ""
    echo "Available services: $(get_services)"
}

case "${1:-}" in
    gen)    gen_mocks "${2:-}" ;;
    clean)  clean_mocks "${2:-}" ;;
    *)      show_help; exit 1 ;;
esac