# Test Directory Structure

This directory contains all testing-related files for the Octopus CLI project.

## Directory Layout

```
test/
├── functional/                 # 功能测试相关文档
│   └── functional-tests.md    # 功能测试用例规范
├── reports/                   # 测试报告
│   └── test-report.md        # 最新的测试执行报告
├── scripts/                  # 测试执行脚本
│   ├── run-functional-tests.sh  # 功能测试执行脚本
│   └── run-e2e-tests.sh         # 端到端生产测试脚本
└── README.md                 # 本文件
```

## File Descriptions

### Functional Tests
- **`functional/functional-tests.md`**: 详细的功能测试用例文档，包含26个测试用例的完整规范
- **`scripts/run-functional-tests.sh`**: 自动化功能测试执行脚本

### End-to-End Tests
- **`scripts/run-e2e-tests.sh`**: 生产级端到端测试，验证代理转发和API切换功能

### Test Reports
- **`reports/test-report.md`**: 最新的功能测试执行报告，包含测试结果分析和质量指标

## Running Tests

### Unit Tests
```bash
# 运行所有单元测试
go test ./...

# 运行特定包的测试
go test ./cmd
go test ./internal/config
go test ./internal/proxy
go test ./internal/process

# 运行测试并显示覆盖率
go test ./... -cover
```

### Functional Tests
```bash
# 从项目根目录运行功能测试
./test/scripts/run-functional-tests.sh
```

### End-to-End Tests
```bash
# 从项目根目录运行端到端测试
./test/scripts/run-e2e-tests.sh
```

### Test Development Mode
```bash
# 使用 TDD watch mode
make tdd

# 运行测试并生成覆盖率报告
make test-coverage
```

## Test Organization Principles

1. **Unit Tests**: 位于各模块的 `*_test.go` 文件中
2. **Functional Tests**: 位于 `test/functional/` 目录
3. **Integration Tests**: 当前通过功能测试覆盖
4. **Test Scripts**: 位于 `scripts/` 目录
5. **Test Reports**: 位于 `test/reports/` 目录

## Test Coverage Goals

- **Unit Test Coverage**: >90%
- **Functional Test Coverage**: 100% (当前已达到)
- **Error Path Coverage**: 95%
- **CLI Command Coverage**: 100%

## Adding New Tests

### Adding Unit Tests
1. 在对应包目录下创建 `*_test.go` 文件
2. 使用 `TestFunction_Scenario_Expected` 命名约定
3. 遵循 TDD 方法论：Red → Green → Refactor

### Adding Functional Tests
1. 在 `functional/functional-tests.md` 中添加测试用例
2. 在 `scripts/run-functional-tests.sh` 中添加测试实现
3. 更新测试报告

### Test Naming Conventions
- **Unit Tests**: `TestFunction_Scenario_Expected`
- **Functional Tests**: `TC-XXX: Description`
- **Test Files**: `*_test.go` for unit tests
- **Test Scripts**: `run-*-tests.sh` for automation scripts