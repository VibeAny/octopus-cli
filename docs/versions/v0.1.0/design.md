# Octopus CLI v0.1.0 设计文档

> **版本**: v0.1.0 - 多代理支持与用户体验优化  
> **状态**: 设计阶段  
> **预计开发周期**: 3-4 个月

## 🎯 版本愿景

将 Octopus CLI 从专用的 Claude Code 工具扩展为支持多种编码代理的通用透传代理，同时大幅提升命令行用户体验。

### 核心价值主张
1. **多代理支持**: 支持 Claude Code、GitHub Codex、Gemini Code 等主流编码代理
2. **智能检测**: 自动识别代理类型，选择对应的 API 配置进行透传
3. **用户体验**: 现代化的命令行界面，美观的进度显示和交互
4. **简单透传**: 保持简单的转发逻辑，不修改请求内容

## 🏗️ 架构设计

### 整体架构

#### 当前架构 (v0.0.4)
```
Claude Code → Octopus CLI → Configured API
```

#### 目标架构 (v0.1.0)
```
Multiple Agents → Agent Detection → Select API from Single Config → Forward Request → Target API
     ↓                ↓                      ↓                         ↓              ↓
[Claude Code]    [User-Agent]         [Single octopus.toml]      [Transparent]   [Anthropic]
[GitHub Codex]   [Header Analysis]    [Multiple APIs]            [Forward]       [OpenAI]  
[Gemini Code]    [Pattern Match]      [Agent Mapping]            [Response]      [Google]
```

### 核心组件重构

#### 1. 代理检测器 (Agent Detector)
**职责**: 识别请求来源的编码代理类型

**检测方法**:
- **User-Agent 解析**: 分析 HTTP User-Agent 头
- **请求头特征**: 检查特定的请求头字段
- **请求路径模式**: 根据 API 路径推断代理类型

**支持的代理**:
```go
type AgentType string

const (
    AgentClaudeCode   AgentType = "claude_code"    // Claude Code
    AgentGitHubCodex  AgentType = "github_codex"   // GitHub Codex/Copilot
    AgentGeminiCode   AgentType = "gemini_code"    // Google Gemini Code
    AgentUnknown      AgentType = "unknown"        // 未知代理，使用默认配置
)

type AgentInfo struct {
    Type        AgentType
    UserAgent   string
    Version     string
    DetectedAt  time.Time
}
```

#### 2. API 选择器 (API Selector)
**职责**: 根据检测到的代理类型从单一配置文件中选择对应的 API

**选择逻辑**:
- 从同一个 octopus.toml 文件中读取所有 API 配置
- 每个 API 配置指定支持的代理类型 
- 根据代理类型匹配对应的 API endpoint
- 支持主备切换（同一代理类型的多个 API 配置）

```go
type APISelector struct {
    config *Config  // 单一配置文件
}

func (as *APISelector) SelectAPIForAgent(agentType AgentType) *APIConfig {
    // 从配置文件的 APIs 列表中查找支持该代理类型的配置
    // 按优先级选择（priority 字段）
    // 支持故障转移到同类型的备用 API
}
```

#### 3. 透传代理 (Transparent Proxy)
**职责**: 完整转发请求和响应，不修改任何内容

```go
type TransparentProxy struct {
    detector  *AgentDetector
    selector  *APISelector    // 改为 APISelector
    httpProxy *httputil.ReverseProxy
}

func (tp *TransparentProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 1. 检测代理类型
    agentInfo := tp.detector.DetectAgent(r)
    
    // 2. 从配置文件中选择对应的 API
    apiConfig := tp.selector.SelectAPIForAgent(agentInfo.Type)
    
    // 3. 透传请求到目标 API
    tp.forwardRequest(w, r, apiConfig)
}
```

## 🎨 用户体验设计

### 视觉设计优化

#### 改进的进度显示
```bash
# 当前版本
⬇️  Downloaded: 50.0% (4.0MB) - 1.2MB/s

# v0.1.0 增强版本  
📦 Downloading octopus v0.1.0...
████████████████████████████████████████ 100%
📊 8.5MB of 8.5MB • 2.1MB/s • Complete ✅
```

#### 美化的状态显示
```bash
# v0.1.0 状态命令
$ octopus status

🐙 Octopus Universal Proxy v0.1.0

📊 Service Status
┌─────────────┬─────────────────────────────┐
│ Status      │ 🟢 Running (PID: 12345)     │
│ Uptime      │ 2h 34m                      │
│ Port        │ 8080                        │
│ Requests    │ 1,247 total                 │
└─────────────┴─────────────────────────────┘

🤖 Supported Agents & APIs
┌─────────────┬────────────────────┬─────────┬───────────┐
│ Agent       │ Target API         │ Status  │ Requests  │
├─────────────┼────────────────────┼─────────┼───────────┤
│ Claude Code │ Anthropic Official │ 🟢 Good │ 856       │
│ GitHub Codex│ OpenAI GPT-4       │ 🟡 Slow │ 234       │
│ Gemini Code │ Google Gemini      │ ❌ Error│ 0         │
└─────────────┴────────────────────┴─────────┴───────────┘

💡 Use 'octopus logs -f' to monitor real-time activity
```

#### 交互式配置
```bash
$ octopus config add

🔧 Add New API Configuration

? Select agent type: 
  > Claude Code (Anthropic)
    GitHub Codex (OpenAI) 
    Gemini Code (Google)

? Configuration name: Anthropic Proxy 1
? API endpoint: https://api.anthropic.com
? API key: sk-ant-******************* ✅ Valid

? Set as primary for Claude Code? (Y/n): Y

✅ Configuration added successfully!
   🤖 Agent: Claude Code
   📝 Name: Anthropic Proxy 1  
   🌐 URL: https://api.anthropic.com
   ⭐ Priority: Primary

💡 Start proxy to use new configuration: octopus start
```

## 🚀 技术实现计划

### 配置文件格式

**v0.1.0 配置文件示例** (单一 octopus.toml):
```toml
[server]
port = 8080
log_level = "info"

[ui]
theme = "modern"           # modern, classic
progress_style = "detailed" # simple, detailed  
colors = true              # 彩色输出开关

# 代理检测配置
[detection]
enabled = true
fallback_agent = "claude_code"  # 检测失败时的默认代理类型

# 代理类型检测规则
[[detection.rules]]
agent = "claude_code"
user_agent_patterns = [
    "^claude-code/",
    "Claude.*Code"
]
headers = ["x-claude-version"]

[[detection.rules]]
agent = "github_codex"
user_agent_patterns = [
    "^github-copilot/",
    "Copilot"
]
headers = ["x-copilot-version"]

[[detection.rules]]
agent = "gemini_code"  
user_agent_patterns = [
    "^gemini-code/",
    "Gemini.*Code"
]

# API 配置 - 在同一个配置文件中为不同代理配置 API
[[apis]]
id = "anthropic_official"
name = "Anthropic Official"
agent_type = "claude_code"    # 指定该 API 支持的代理类型
url = "https://api.anthropic.com"
api_key = "sk-xxx"
priority = 1                 # 1=主要，2=备用
enabled = true

[[apis]]
id = "anthropic_proxy"
name = "Anthropic Proxy"
agent_type = "claude_code"    # 同一代理类型的备用 API
url = "https://proxy.anthropic.com"
api_key = "pk-xxx"
priority = 2                 # 备用优先级
enabled = true

[[apis]]
id = "openai_official"
name = "OpenAI Official"
agent_type = "github_codex"   # 为 GitHub Codex 配置的 API
url = "https://api.openai.com"
api_key = "sk-xxx"
priority = 1
enabled = true

[[apis]]
id = "google_gemini"
name = "Google Gemini"
agent_type = "gemini_code"    # 为 Gemini Code 配置的 API
url = "https://generativelanguage.googleapis.com"
api_key = "AIza-xxx"
priority = 1
enabled = false              # 暂时禁用

[settings]
# 全局设置保持兼容 v0.0.x
log_file = "logs/octopus.log"
```

**配置文件管理**:
- ✅ 保持使用单一的 `octopus.toml` 配置文件
- ✅ 在同一个文件中为不同代理类型配置不同的 API
- ✅ 每个 API 配置通过 `agent_type` 字段指定支持的代理
- ✅ 支持同一代理类型配置多个 API（主备关系）
- ✅ 完全向后兼容现有的 v0.0.x 配置格式

### 开发阶段规划

#### Phase 13: 代理检测基础 (2-3 周)
- [ ] 实现 User-Agent 解析器
- [ ] 支持请求头特征检测
- [ ] 建立代理类型注册机制
- [ ] 实现检测规则配置化

#### Phase 14: 配置管理扩展 (2-3 周)  
- [ ] 扩展现有配置文件格式，添加 `agent_type` 字段支持
- [ ] 实现基于代理类型的 API 选择逻辑
- [ ] 保持单一配置文件，支持多代理 API 配置  
- [ ] 添加配置验证和健康检查
- [ ] 提供配置迁移工具（从 v0.0.x 平滑升级）

#### Phase 15: 用户体验优化 (3-4 周)
- [ ] 美化所有命令的输出格式
- [ ] 实现进度条和 spinner 动画
- [ ] 添加交互式配置向导
- [ ] 优化错误信息显示和建议

#### Phase 16: 核心功能集成 (2-3 周)
- [ ] 集成代理检测到透传代理
- [ ] 实现基于代理的配置选择
- [ ] 添加请求统计和监控
- [ ] 完善日志记录和调试信息

#### Phase 17: 测试和优化 (2-3 周)
- [ ] 多代理兼容性测试
- [ ] 用户体验测试和改进
- [ ] 性能测试和优化
- [ ] 文档更新和示例完善

#### Phase 18: 发布准备 (1-2 周)
- [ ] Beta 版本发布
- [ ] 用户反馈收集和问题修复
- [ ] 最终测试和发布准备
- [ ] v0.1.0 正式发布

## 📝 命令接口扩展

### 新增和改进的命令

```bash
# 代理相关命令
octopus agents list              # 列出支持的代理类型
octopus agents detect           # 测试代理检测功能
octopus agents stats            # 显示各代理使用统计

# 配置管理增强
octopus config add --agent=claude_code     # 为特定代理添加配置
octopus config list --agent=all            # 按代理分组显示配置
octopus config test --agent=github_codex   # 测试特定代理的配置

# 增强的监控命令
octopus status --live           # 实时状态监控
octopus health --verbose        # 详细健康检查
octopus metrics                 # 显示使用指标
```

## 🎯 成功标准

### 功能目标
- ✅ 支持至少 3 种主流编码代理 (Claude Code, GitHub Codex, Gemini Code)
- ✅ 代理检测准确率 > 95%
- ✅ 100% 向后兼容 v0.0.x 配置
- ✅ 配置切换延迟 < 100ms

### 用户体验目标
- ✅ 所有命令输出美观统一
- ✅ 进度显示清晰直观
- ✅ 错误信息有帮助性
- ✅ 交互流程简单易用

### 技术质量目标
- ✅ 单元测试覆盖率 > 90%
- ✅ 透传延迟增加 < 5ms
- ✅ 内存占用增加 < 20MB
- ✅ 支持所有现有平台

这个简化版本专注于核心需求：多代理支持和用户体验优化，避免了过度工程化，保持了系统的简洁性和可维护性。