# Octopus CLI 功能性测试项目

## 测试环境准备

### 前置条件
- [x] Go 项目构建成功 (`go build -o octopus ./cmd`)
- [ ] 创建测试配置目录
- [ ] 准备测试用的 API 配置
- [ ] 清理之前的测试残留

### 测试数据准备
```bash
# 测试配置文件路径
TEST_CONFIG_DIR="./test-configs"
TEST_CONFIG_FILE="./test-configs/octopus-test.toml"

# 测试用的 API 配置 (不包含真实密钥)
API_1_NAME="test-api-1"
API_1_URL="https://httpbin.org/anything"
API_1_KEY="test-key-1"

API_2_NAME="test-api-2" 
API_2_URL="https://jsonplaceholder.typicode.com"
API_2_KEY="test-key-2"
```

## 测试用例清单

### 1. 基础命令测试 (Basic Commands)

#### TC-001: 帮助信息测试
- **测试目标**: 验证帮助信息显示正确
- **测试步骤**:
  1. 执行 `./octopus --help`
  2. 执行 `./octopus -h`
  3. 执行 `./octopus`
- **预期结果**: 显示完整的命令帮助信息，包含所有子命令

#### TC-002: 版本信息测试
- **测试目标**: 验证版本信息显示正确
- **测试步骤**:
  1. 执行 `./octopus version`
  2. 执行 `./octopus --version`
- **预期结果**: 显示正确的版本号 (dev)

### 2. 配置管理命令测试 (Config Management)

#### TC-003: 配置列表测试 (空配置)
- **测试目标**: 验证空配置时的列表显示
- **测试步骤**:
  1. 清理配置文件
  2. 执行 `./octopus config list`
- **预期结果**: 显示 "No APIs configured" 消息

#### TC-004: 添加配置测试
- **测试目标**: 验证添加新配置功能
- **测试步骤**:
  1. 执行 `./octopus config add test-api-1 https://httpbin.org/anything test-key-1`
  2. 执行 `./octopus config list`
- **预期结果**: 
  - 添加成功消息
  - 列表中显示新添加的配置

#### TC-005: 添加重复配置测试
- **测试目标**: 验证重复配置的错误处理
- **测试步骤**:
  1. 重复执行 `./octopus config add test-api-1 https://httpbin.org/anything test-key-1`
- **预期结果**: 显示"已存在"错误信息

#### TC-006: 添加第二个配置
- **测试目标**: 验证多个配置管理
- **测试步骤**:
  1. 执行 `./octopus config add test-api-2 https://jsonplaceholder.typicode.com test-key-2`
  2. 执行 `./octopus config list`
- **预期结果**: 列表显示两个配置

#### TC-007: 显示配置详情测试
- **测试目标**: 验证配置详情显示
- **测试步骤**:
  1. 执行 `./octopus config show test-api-1`
  2. 执行 `./octopus config show test-api-2`
- **预期结果**: 显示配置的详细信息，API密钥被掩码处理

#### TC-008: 显示不存在配置测试
- **测试目标**: 验证不存在配置的错误处理
- **测试步骤**:
  1. 执行 `./octopus config show nonexistent`
- **预期结果**: 显示"配置未找到"错误信息

#### TC-009: 切换配置测试
- **测试目标**: 验证配置切换功能
- **测试步骤**:
  1. 执行 `./octopus config switch test-api-1`
  2. 执行 `./octopus config list`
  3. 执行 `./octopus config switch test-api-2`
  4. 执行 `./octopus config list`
- **预期结果**: 
  - 切换成功消息
  - 列表中显示正确的 [ACTIVE] 标记

#### TC-010: 切换到不存在配置测试
- **测试目标**: 验证切换不存在配置的错误处理
- **测试步骤**:
  1. 执行 `./octopus config switch nonexistent`
- **预期结果**: 显示"配置未找到"错误信息

#### TC-011: 删除配置测试
- **测试目标**: 验证删除配置功能
- **测试步骤**:
  1. 执行 `./octopus config remove test-api-2`
  2. 执行 `./octopus config list`
- **预期结果**: 
  - 删除成功消息
  - 列表中不再显示被删除的配置

#### TC-012: 删除激活配置测试
- **测试目标**: 验证删除当前激活配置的处理
- **测试步骤**:
  1. 确保 test-api-1 是激活状态
  2. 执行 `./octopus config remove test-api-1`
  3. 执行 `./octopus config list`
- **预期结果**: 
  - 删除成功消息
  - 显示"清除激活API"消息
  - 列表为空

#### TC-013: 删除不存在配置测试
- **测试目标**: 验证删除不存在配置的错误处理
- **测试步骤**:
  1. 执行 `./octopus config remove nonexistent`
- **预期结果**: 显示"配置未找到"错误信息

### 3. 服务管理命令测试 (Service Management)

#### TC-014: 状态查询测试 (初始状态)
- **测试目标**: 验证初始服务状态
- **测试步骤**:
  1. 执行 `./octopus status`
- **预期结果**: 显示服务未运行状态

#### TC-015: 启动服务测试 (无配置)
- **测试目标**: 验证无配置时的启动处理
- **测试步骤**:
  1. 确保没有配置
  2. 执行 `./octopus start`
- **预期结果**: 可能显示警告但不应崩溃

#### TC-016: 启动服务测试 (有配置)
- **测试目标**: 验证有配置时的服务启动
- **测试步骤**:
  1. 添加测试配置: `./octopus config add test-api https://httpbin.org/anything test-key`
  2. 切换配置: `./octopus config switch test-api`
  3. 执行 `./octopus start`
  4. 执行 `./octopus status`
- **预期结果**: 
  - 启动成功消息
  - 状态显示服务正在运行

#### TC-017: 重复启动服务测试
- **测试目标**: 验证重复启动的错误处理
- **测试步骤**:
  1. 在服务运行时执行 `./octopus start`
- **预期结果**: 显示"服务已在运行"错误信息

#### TC-018: 停止服务测试
- **测试目标**: 验证服务停止功能
- **测试步骤**:
  1. 执行 `./octopus stop`
  2. 执行 `./octopus status`
- **预期结果**: 
  - 停止成功消息
  - 状态显示服务已停止

#### TC-019: 停止未运行服务测试
- **测试目标**: 验证停止未运行服务的错误处理
- **测试步骤**:
  1. 确保服务未运行
  2. 执行 `./octopus stop`
- **预期结果**: 显示"服务未运行"错误信息

### 4. 监控诊断命令测试 (Monitoring & Diagnostics)

#### TC-020: 健康检查测试 (无配置)
- **测试目标**: 验证无配置时的健康检查
- **测试步骤**:
  1. 清理所有配置
  2. 执行 `./octopus health`
- **预期结果**: 显示"无API配置"消息

#### TC-021: 健康检查测试 (有配置)
- **测试目标**: 验证有配置时的健康检查
- **测试步骤**:
  1. 添加测试配置
  2. 执行 `./octopus health`
- **预期结果**: 显示各API配置的健康状态

#### TC-022: 日志查看测试 (无日志文件)
- **测试目标**: 验证无日志文件时的处理
- **测试步骤**:
  1. 执行 `./octopus logs`
- **预期结果**: 显示"日志文件未找到"错误信息

#### TC-023: 日志查看测试 (有日志文件)
- **测试目标**: 验证日志文件存在时的显示
- **测试步骤**:
  1. 创建测试日志文件
  2. 执行 `./octopus logs`
- **预期结果**: 显示日志内容

#### TC-024: 日志follow标志测试
- **测试目标**: 验证 --follow 标志
- **测试步骤**:
  1. 执行 `./octopus logs --follow` (快速取消)
- **预期结果**: 显示follow模式提示

### 5. 配置文件测试 (Configuration Files)

#### TC-025: 自定义配置文件测试
- **测试目标**: 验证 -f 参数指定配置文件
- **测试步骤**:
  1. 执行 `./octopus -f ./test-configs/custom.toml config list`
- **预期结果**: 使用指定的配置文件

#### TC-026: 不存在配置文件测试
- **测试目标**: 验证不存在配置文件的处理
- **测试步骤**:
  1. 执行 `./octopus -f /nonexistent/config.toml config list`
- **预期结果**: 创建默认配置或显示合适错误

#### TC-027: 无效配置文件测试
- **测试目标**: 验证无效TOML文件的错误处理
- **测试步骤**:
  1. 创建无效TOML文件
  2. 执行 `./octopus -f ./invalid.toml config list`
- **预期结果**: 显示配置文件解析错误

### 6. 边界条件和错误处理测试 (Edge Cases & Error Handling)

#### TC-028: 无效参数测试
- **测试目标**: 验证无效命令参数的处理
- **测试步骤**:
  1. 执行 `./octopus invalid-command`
  2. 执行 `./octopus config invalid-subcommand`
  3. 执行 `./octopus config add` (缺少参数)
- **预期结果**: 显示合适的错误信息和帮助提示

#### TC-029: 权限测试
- **测试目标**: 验证文件权限问题的处理
- **测试步骤**:
  1. 创建只读配置目录
  2. 尝试添加配置
- **预期结果**: 显示权限错误信息

#### TC-030: 长参数测试
- **测试目标**: 验证长参数值的处理
- **测试步骤**:
  1. 添加很长的URL和API密钥
- **预期结果**: 正确处理长参数值

### 7. 集成测试 (Integration Tests)

#### TC-031: 完整工作流测试
- **测试目标**: 验证完整的使用流程
- **测试步骤**:
  1. 添加多个配置
  2. 启动服务
  3. 切换配置
  4. 检查健康状态
  5. 查看日志
  6. 停止服务
- **预期结果**: 整个流程正常工作

#### TC-032: 并发操作测试
- **测试目标**: 验证并发操作的安全性
- **测试步骤**:
  1. 同时执行多个配置操作
- **预期结果**: 操作原子性，无数据损坏

## 测试执行计划

### 执行顺序
1. **环境准备** → 清理环境，构建项目
2. **基础功能** → TC-001 到 TC-002
3. **配置管理** → TC-003 到 TC-013
4. **服务管理** → TC-014 到 TC-019
5. **监控诊断** → TC-020 到 TC-024
6. **配置文件** → TC-025 到 TC-027
7. **边界条件** → TC-028 到 TC-030
8. **集成测试** → TC-031 到 TC-032

### 测试报告格式
```
TC-XXX: [测试名称]
状态: [PASS/FAIL/SKIP]
执行时间: [timestamp]
结果: [实际结果]
备注: [问题说明或改进建议]
```

## 测试通过标准

### 必须通过的测试
- 所有基础命令测试 (TC-001, TC-002)
- 核心配置管理功能 (TC-004, TC-006, TC-007, TC-009, TC-011)
- 基本服务管理功能 (TC-016, TC-018)
- 错误处理测试 (TC-005, TC-010, TC-013, TC-017, TC-019)

### 可接受的失败
- 部分监控功能测试 (健康检查具体实现)
- 权限相关测试 (依赖系统环境)
- 并发测试 (复杂性较高)

## 后续优化建议
基于测试结果，标记需要改进的功能点：
- [ ] 用户体验优化点
- [ ] 性能优化点  
- [ ] 错误信息改进点
- [ ] 功能增强点