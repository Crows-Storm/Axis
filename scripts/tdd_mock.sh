#!/bin/bash
set -e

# 终端颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 默认 Mock 输出目录
MOCK_DIR="./mocks"

# 检查 mockery 是否安装
check_mockery() {
    if ! command -v mockery &> /dev/null; then
        echo -e "${RED}❌ Error: mockery is not installed.${NC}"
        echo -e "${YELLOW}💡 Please install it using:${NC}"
        echo "   go install github.com/vektra/mockery/v2@latest"
        exit 1
    fi
}

# 生成 Mock 代码
gen_mocks() {
    check_mockery
    echo -e "${GREEN}🚀 Generating mocks...${NC}"

    # 如果存在 .mockery.yaml 配置文件，优先使用配置文件 (推荐做法)
    if [ -f ".mockery.yaml" ]; then
        echo -e "${YELLOW}📄 Using .mockery.yaml configuration...${NC}"
        mockery
    else
        # 降级方案：如果没有配置文件，使用命令行参数扫描 internal/ 目录
        echo -e "${YELLOW}⚠️ No .mockery.yaml found. Using default CLI arguments...${NC}"
        mockery --dir=./internal --output=$MOCK_DIR --outpkg=mocks --all
    fi

    echo -e "${GREEN}✅ Mocks generated successfully in ${MOCK_DIR}${NC}"
}

# 清理 Mock 代码
clean_mocks() {
    echo -e "${YELLOW}🧹 Cleaning generated mocks in ${MOCK_DIR}...${NC}"

    if [ -d "$MOCK_DIR" ]; then
        rm -rf "$MOCK_DIR"
        echo -e "${GREEN}✅ Mocks cleaned successfully.${NC}"
    else
        echo -e "${YELLOW}⚠️ Directory ${MOCK_DIR} does not exist. Nothing to clean.${NC}"
    fi
}

show_help() {
    echo "Usage: $0 {gen|clean}"
    echo ""
    echo "Commands:"
    echo "  gen    Generate mock files using mockery"
    echo "  clean  Remove all generated mock files"
}

# 主入口
case "$1" in
    gen)
        gen_mocks
        ;;
    clean)
        clean_mocks
        ;;
    *)
        show_help
        exit 1
        ;;
esac