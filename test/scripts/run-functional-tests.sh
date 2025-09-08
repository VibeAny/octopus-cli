#!/bin/bash

# Octopus CLI Functional Test Runner
# 功能性测试执行脚本
# 
# Usage: ./scripts/run-functional-tests.sh
# 
# This script should be run from the project root directory

set -e  # 遇到错误时停止

# 确保从项目根目录运行
if [[ ! -f "go.mod" ]]; then
    echo "❌ Error: This script must be run from the project root directory"
    echo "Usage: ./scripts/run-functional-tests.sh"
    exit 1
fi

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 测试统计
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# 测试配置
TEST_CONFIG_DIR="./test-configs"
TEST_CONFIG_FILE="$TEST_CONFIG_DIR/octopus-test.toml"
OCTOPUS_BIN="./octopus"

# 创建测试目录
mkdir -p "$TEST_CONFIG_DIR"

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
}

# 测试执行函数
run_test() {
    local test_id="$1"
    local test_name="$2"
    local expected="$3"
    shift 3
    local command="$@"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo ""
    echo "========================================"
    echo "$test_id: $test_name"
    echo "========================================"
    echo "Command: $command"
    echo "Expected: $expected"
    echo "----------------------------------------"
    
    local start_time=$(date +%s)
    local output
    local exit_code
    
    # 执行命令并捕获输出和退出码
    if output=$(eval "$command" 2>&1); then
        exit_code=0
    else
        exit_code=$?
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo "Output:"
    echo "$output"
    echo "Exit code: $exit_code"
    echo "Duration: ${duration}s"
    
    # 检查结果 (简化版本，实际应该更智能地匹配预期结果)
    if [[ "$output" =~ "$expected" ]] || [[ "$expected" == "ANY" ]]; then
        log_success "$test_id PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        log_error "$test_id FAILED"
        log_error "Expected to contain: $expected"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

# 跳过测试函数
skip_test() {
    local test_id="$1"
    local test_name="$2"
    local reason="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    SKIPPED_TESTS=$((SKIPPED_TESTS + 1))
    log_skip "$test_id: $test_name (Reason: $reason)"
}

# 清理函数
cleanup() {
    log_info "Cleaning up test environment..."
    
    # 停止任何运行中的服务
    $OCTOPUS_BIN stop 2>/dev/null || true
    
    # 删除测试配置目录
    rm -rf "$TEST_CONFIG_DIR" 2>/dev/null || true
    
    # 删除可能的PID文件
    rm -f /tmp/octopus.pid 2>/dev/null || true
    
    # 删除默认配置目录的测试数据
    rm -rf ~/.config/octopus 2>/dev/null || true
    
    log_info "Cleanup completed"
}

# 环境准备
prepare_environment() {
    log_info "Preparing test environment..."
    
    # 清理之前的测试残留
    cleanup
    
    # 重新创建测试目录
    mkdir -p "$TEST_CONFIG_DIR"
    
    # 检查octopus二进制是否存在
    if [[ ! -f "$OCTOPUS_BIN" ]]; then
        log_error "Octopus binary not found at $OCTOPUS_BIN"
        log_info "Building octopus binary..."
        go build -o octopus ./cmd
        if [[ $? -ne 0 ]]; then
            log_error "Failed to build octopus binary"
            exit 1
        fi
    fi
    
    log_success "Environment prepared successfully"
}

# 生成测试报告
generate_report() {
    echo ""
    echo "========================================"
    echo "TEST EXECUTION SUMMARY"
    echo "========================================"
    echo "Total Tests: $TOTAL_TESTS"
    echo "Passed: $PASSED_TESTS"
    echo "Failed: $FAILED_TESTS"
    echo "Skipped: $SKIPPED_TESTS"
    echo ""
    
    if [[ $FAILED_TESTS -eq 0 ]]; then
        log_success "All tests passed! ✨"
    else
        log_error "$FAILED_TESTS test(s) failed!"
    fi
    
    local success_rate=$((PASSED_TESTS * 100 / (TOTAL_TESTS - SKIPPED_TESTS)))
    echo "Success Rate: $success_rate%"
}

# 主测试执行函数
main() {
    echo "🚀 Starting Octopus CLI Functional Tests"
    echo ""
    
    # 准备环境
    prepare_environment
    
    # 1. 基础命令测试
    echo ""
    log_info "=== SECTION 1: Basic Commands ==="
    
    run_test "TC-001" "Help information test" "Octopus CLI" \
        "$OCTOPUS_BIN --help"
    
    run_test "TC-002" "Version information test" "version dev" \
        "$OCTOPUS_BIN version"
    
    # 2. 配置管理命令测试
    echo ""
    log_info "=== SECTION 2: Configuration Management ==="
    
    run_test "TC-003" "Empty config list test" "No APIs configured" \
        "$OCTOPUS_BIN config list"
    
    run_test "TC-004" "Add configuration test" "Added API configuration" \
        "$OCTOPUS_BIN config add test-api-1 https://httpbin.org/anything test-key-1"
    
    run_test "TC-005" "Duplicate configuration test" "already exists" \
        "$OCTOPUS_BIN config add test-api-1 https://httpbin.org/anything test-key-1"
    
    run_test "TC-006" "Add second configuration" "Added API configuration" \
        "$OCTOPUS_BIN config add test-api-2 https://jsonplaceholder.typicode.com test-key-2"
    
    run_test "TC-007" "Show configuration details" "API Configuration: test-api-1" \
        "$OCTOPUS_BIN config show test-api-1"
    
    run_test "TC-008" "Show nonexistent configuration" "not found" \
        "$OCTOPUS_BIN config show nonexistent"
    
    run_test "TC-009" "Switch configuration test" "Switched to API" \
        "$OCTOPUS_BIN config switch test-api-1"
    
    run_test "TC-010" "Switch to nonexistent config" "not found" \
        "$OCTOPUS_BIN config switch nonexistent"
    
    run_test "TC-011" "Remove configuration test" "Removed API configuration" \
        "$OCTOPUS_BIN config remove test-api-2"
    
    run_test "TC-012" "Remove active configuration" "Cleared active API" \
        "$OCTOPUS_BIN config remove test-api-1"
    
    run_test "TC-013" "Remove nonexistent config" "not found" \
        "$OCTOPUS_BIN config remove nonexistent"
    
    # 重新添加配置用于服务测试
    $OCTOPUS_BIN config add test-api https://httpbin.org/anything test-key >/dev/null 2>&1 || true
    $OCTOPUS_BIN config switch test-api >/dev/null 2>&1 || true
    
    # 3. 服务管理命令测试
    echo ""
    log_info "=== SECTION 3: Service Management ==="
    
    run_test "TC-014" "Initial service status" "Status: Stopped" \
        "$OCTOPUS_BIN status"
    
    run_test "TC-016" "Start service test" "Service started successfully" \
        "$OCTOPUS_BIN start"
    
    sleep 2  # 给服务一些启动时间
    
    run_test "TC-017" "Duplicate start test" "already running" \
        "$OCTOPUS_BIN start"
    
    run_test "TC-018" "Stop service test" "Service stopped successfully" \
        "$OCTOPUS_BIN stop"
    
    run_test "TC-019" "Stop non-running service" "not running" \
        "$OCTOPUS_BIN stop"
    
    # 4. 监控诊断命令测试
    echo ""
    log_info "=== SECTION 4: Monitoring & Diagnostics ==="
    
    # 清理配置进行健康检查测试
    cleanup
    run_test "TC-020" "Health check without config" "No APIs configured" \
        "$OCTOPUS_BIN health"
    
    # 重新添加配置
    $OCTOPUS_BIN config add test-api https://httpbin.org/anything test-key >/dev/null 2>&1 || true
    
    run_test "TC-021" "Health check with config" "Checking API endpoints health" \
        "$OCTOPUS_BIN health"
    
    run_test "TC-022" "Logs test without log file" "log file not found" \
        "$OCTOPUS_BIN logs"
    
    # 5. 配置文件测试
    echo ""
    log_info "=== SECTION 5: Configuration Files ==="
    
    run_test "TC-025" "Custom config file test" "ANY" \
        "$OCTOPUS_BIN -f $TEST_CONFIG_FILE config list"
    
    run_test "TC-026" "Nonexistent config file" "ANY" \
        "$OCTOPUS_BIN -f /nonexistent/config.toml config list"
    
    # 6. 边界条件测试
    echo ""
    log_info "=== SECTION 6: Edge Cases & Error Handling ==="
    
    run_test "TC-028a" "Invalid command test" "Error" \
        "$OCTOPUS_BIN invalid-command"
    
    run_test "TC-028b" "Invalid subcommand test" "Error" \
        "$OCTOPUS_BIN config invalid-subcommand"
    
    run_test "TC-028c" "Missing parameters test" "Error" \
        "$OCTOPUS_BIN config add"
    
    # 清理
    cleanup
    
    # 生成报告
    generate_report
    
    # 根据结果返回适当的退出码
    if [[ $FAILED_TESTS -eq 0 ]]; then
        exit 0
    else
        exit 1
    fi
}

# 捕获中断信号进行清理
trap cleanup EXIT INT TERM

# 执行主函数
main "$@"