# Octopus CLI - Claude Code API 转发服务需求文档

## 项目背景

Claude Code API 切换的痛点：
- 用户需要频繁切换不同的 API 提供商（官方API、第三方代理等）
- 每次切换需要修改环境变量
- 修改后需要重启 Claude Code 才能生效
- 操作繁琐，影响开发效率

## 项目目标

开发一个命令行工具 (CLI)，提供本地 API 转发服务，实现：
1. 统一的 API 转发入口
2. 通过命令行动态切换多个 API 配置
3. 无需重启即可切换 API
4. 简化用户操作流程

## 技术选型

- **开发语言**: Go
- **架构模式**: CLI + HTTP 代理服务
- **配置格式**: TOML
- **部署方式**: 本地命令行工具

## 功能需求

### 核心功能

1. **CLI 命令接口**
   - 服务管理命令（启动、停止、状态查询）
   - 配置管理命令（列表、添加、删除、切换）
   - 健康检查命令

2. **API 转发代理**
   - 接收来自 Claude Code 的请求
   - 根据当前配置转发到目标 API 服务
   - 处理请求/响应的转发和格式化

3. **TOML 配置管理**
   - 支持配置多个 API 端点
   - 每个配置包含：ID、名称、URL、API Key、状态
   - 配置持久化存储到 TOML 文件

4. **动态切换**
   - 提供 CLI 命令切换当前使用的 API 配置
   - 实时生效，无需重启服务
   - 支持通过命令行或管理 API 切换

5. **健康检查**
   - 检查各个 API 端点的可用性
   - 提供状态监控接口

### CLI 命令设计

1. **服务管理**
   ```bash
   octopus start          # 启动代理服务
   octopus stop           # 停止代理服务
   octopus status         # 查看服务状态
   octopus restart        # 重启服务
   ```

2. **配置管理**
   ```bash
   octopus config list                    # 列出所有配置
   octopus config add <name> <url> <key>  # 添加新配置
   octopus config remove <id>             # 删除配置
   octopus config switch <id>             # 切换到指定配置
   octopus config show <id>               # 显示配置详情
   ```

3. **监控和诊断**
   ```bash
   octopus health         # 检查所有 API 健康状态
   octopus logs           # 查看服务日志
   octopus version        # 显示版本信息
   ```

### 管理功能

1. **配置管理 API**
   - 添加/删除/修改 API 配置
   - 查看当前配置列表
   - 获取当前活跃配置

2. **状态监控**
   - 服务运行状态
   - API 调用统计
   - 错误日志记录

## 技术架构

### 组件设计

1. **CLI 接口**
   - 命令解析和路由
   - 用户交互界面

2. **HTTP 代理服务器**
   - 监听本地端口（如 :8080）
   - 处理 Claude Code 的 API 请求

3. **TOML 配置管理器**
   - 管理 API 配置的 CRUD 操作
   - TOML 配置文件读写

4. **转发引擎**
   - 请求转发逻辑
   - 响应处理

5. **管理接口**
   - RESTful API 用于配置管理
   - CLI 命令后端支持

### 数据结构

```go
type APIConfig struct {
    ID       string `toml:"id"`
    Name     string `toml:"name"`
    URL      string `toml:"url"`
    APIKey   string `toml:"api_key"`
    IsActive bool   `toml:"is_active"`
}

type ServerConfig struct {
    Port     int    `toml:"port"`
    LogLevel string `toml:"log_level"`
}

type Settings struct {
    ActiveAPI string `toml:"active_api"`
}

type Config struct {
    Server   ServerConfig `toml:"server"`
    APIs     []APIConfig  `toml:"apis"`
    Settings Settings     `toml:"settings"`
}
```

## TOML 配置文件设计

### 主配置文件 (octopus.toml)
```toml
[server]
port = 8080
log_level = "info"

[[apis]]
id = "official"
name = "Anthropic Official"
url = "https://api.anthropic.com"
api_key = "sk-xxx"
is_active = true

[[apis]]
id = "proxy1"
name = "Proxy Service 1"
url = "https://api.proxy1.com"
api_key = "pk-xxx"
is_active = false

[settings]
active_api = "official"
```

## 用户使用流程

1. **初始设置**
   ```bash
   octopus config add official https://api.anthropic.com sk-xxx
   octopus config add proxy1 https://api.proxy1.com pk-xxx
   octopus start
   ```

2. **设置 Claude Code**
   - 配置 Claude Code 使用本地代理地址 `http://localhost:8080`

3. **日常使用**
   ```bash
   octopus config list              # 查看所有配置
   octopus config switch proxy1     # 切换到 proxy1
   octopus status                   # 检查服务状态
   ```

## 非功能需求

1. **性能**
   - 低延迟转发
   - 支持并发请求

2. **可靠性**
   - 服务稳定性
   - 错误处理和恢复
   - 配置备份和恢复

3. **易用性**
   - 直观的 CLI 命令
   - 清晰的错误提示
   - 完善的帮助文档

4. **安全性**
   - API Key 安全存储
   - 请求日志脱敏
   - 配置文件权限控制