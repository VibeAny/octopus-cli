# Octopus CLI 项目任务管理

## 开发阶段划分

### Phase 1: 项目基础设置 (已完成)
- [x] 创建项目目录结构
- [x] 编写需求文档 (CLI 版本)
- [x] 建立任务管理体系
- [x] 更新为 CLI 架构设计

### Phase 2: Go CLI 项目初始化 (已完成)
- [x] 初始化 Git 仓库和开源准备
  - [x] `git init` 初始化仓库
  - [x] 创建 MIT License 文件
  - [x] 创建 .gitignore 文件
  - [x] 创建 README.md 项目介绍
- [x] 初始化 Go 模块 (`go mod init octopus-cli`)
- [x] 创建 CLI 项目结构
  - [x] `cmd/` - CLI 入口和命令 (简化为单文件结构)
  - [x] `internal/` - 内部包
  - [x] `pkg/` - 公共包  
  - [x] `configs/` - TOML 配置文件
- [x] 集成 Cobra CLI 框架
- [x] 设置基本的 Makefile
- [x] 设置 TDD 测试框架

### Phase 3: CLI 基础架构实现 (TDD 方式) (已完成)
- [x] 实现 CLI 命令结构 (测试先行)
  - [x] 编写根命令和子命令测试
  - [x] 根命令和子命令定义
  - [x] 编写命令行参数解析测试  
  - [x] 命令行参数解析
  - [x] 编写帮助信息测试
  - [x] 帮助信息和版本信息
- [x] 实现 TOML 配置管理器 (测试先行)
  - [x] 编写配置数据结构测试
  - [x] 定义数据结构
  - [x] 编写 TOML 配置文件读写测试
  - [x] TOML 配置文件读写
  - [x] 编写配置验证测试
  - [x] 配置验证逻辑
- [x] 实现基础服务管理 (测试先行)
  - [x] 编写进程管理器测试
  - [x] 进程管理器
  - [x] 编写守护进程测试
  - [x] 守护进程支持
  - [x] 编写 PID 文件管理测试
  - [x] PID 文件管理

### Phase 4: 核心代理功能实现 (TDD 方式) (已完成)
- [x] 实现 HTTP 代理服务器 (测试先行) (已完成)
  - [x] 编写基础 HTTP 服务器测试
  - [x] 基础 HTTP 服务器
  - [x] 编写请求拦截和转发测试
  - [x] 请求拦截和转发
  - [x] 编写响应处理测试
  - [x] 响应处理
- [x] 实现转发引擎 (测试先行) (已完成)
  - [x] 编写 API 请求转发逻辑测试
  - [x] API 请求转发逻辑
  - [x] 编写错误处理和重试测试
  - [x] 错误处理和重试
  - [x] 编写日志记录测试
  - [x] 日志记录
- [x] 实现配置切换功能 (测试先行) (已完成)
  - [x] 编写动态配置重载测试
  - [x] 动态配置重载
  - [x] 编写活跃配置管理测试
  - [x] 活跃配置管理

### Phase 5: CLI 命令实现 (TDD 方式) (已完成)
- [x] 服务管理命令 (测试先行)
  - [x] 编写 `octopus start` 测试 - 启动服务
  - [x] `octopus start` - 启动服务
  - [x] 编写 `octopus stop` 测试 - 停止服务
  - [x] `octopus stop` - 停止服务  
  - [x] 编写 `octopus status` 测试 - 状态查询
  - [x] `octopus status` - 状态查询
  - [ ] 编写 `octopus restart` 测试 - 重启服务 (不在 MVP 范围)
  - [ ] `octopus restart` - 重启服务 (不在 MVP 范围)
- [x] 配置管理命令 (测试先行)
  - [x] 编写 `octopus config list` 测试 - 列出配置
  - [x] `octopus config list` - 列出配置
  - [x] 编写 `octopus config add` 测试 - 添加配置
  - [x] `octopus config add` - 添加配置
  - [x] 编写 `octopus config remove` 测试 - 删除配置
  - [x] `octopus config remove` - 删除配置
  - [x] 编写 `octopus config switch` 测试 - 切换配置
  - [x] `octopus config switch` - 切换配置
  - [x] 编写 `octopus config show` 测试 - 显示配置详情
  - [x] `octopus config show` - 显示配置详情
  - [ ] 编写 `octopus config edit` 测试 - 编辑配置文件 (新增需求)
  - [ ] `octopus config edit` - 使用系统默认编辑器打开当前配置文件 (新增需求)
- [x] 监控诊断命令 (测试先行)
  - [x] 编写 `octopus health` 测试 - 健康检查
  - [x] `octopus health` - 健康检查
  - [x] 编写 `octopus logs` 测试 - 日志查看
  - [x] `octopus logs` - 日志查看
  - [x] 编写 `octopus version` 测试 - 版本信息
  - [x] `octopus version` - 版本信息

### Phase 6: 用户体验优化 (已完成核心功能)
- [x] CLI 用户界面优化
  - [x] 便捷启动模式 (`octopus` 无参数直接启动)
  - [x] 智能配置文件管理 (自动使用 default.toml 或记住上次使用的配置)
  - [x] 状态查看功能 (`octopus config` 显示当前配置信息)
  - [x] 彩色输出和格式化 (已完成表格对齐修复)
  - [ ] 进度指示器
- [x] 错误处理优化
  - [x] 用户友好的错误信息
  - [x] 配置文件不存在时自动回退到默认配置
  - [x] API 切换时自动验证 API 配置存在
- [x] 配置管理优化
  - [x] 可移植配置系统 (所有文件保存在二进制同目录)
  - [x] 配置状态持久化 (settings.toml 记录当前配置文件)
  - [x] API 即时切换 (自动重启守护进程应用新配置)
- [x] 构建系统优化
  - [x] 多平台构建支持 (Windows, macOS-x64, macOS-ARM64, Linux)
  - [x] 版本化命名规则 (v0.0.1-platform-YYYYMMDD.git_sha)

### Phase 7: 健康检查和监控 (已完成基础功能)
- [x] API 健康检查功能
  - [x] 端点可用性检查 (`octopus health` 命令)
  - [x] 响应时间监控
  - [x] 状态报告生成
- [x] 日志和监控系统
  - [x] 结构化日志 (时间戳 + 日志级别 + 消息)
  - [x] 请求转发日志 (记录源 IP、目标 API、转发结果)
  - [x] 配置变更日志 (记录 API 切换和守护进程重启)
  - [x] 可移植日志文件 (logs/octopus.log)
  - [ ] 请求统计
  - [ ] 性能指标收集

### Phase 8: 用户体验增强功能 (已完成)
- [x] 配置文件编辑功能 (测试先行)
  - [x] 编写跨平台编辑器检测测试
  - [x] 系统默认编辑器检测逻辑
    - [x] Linux/macOS: 优先使用 `$EDITOR` 环境变量
    - [x] 如果未设置，按优先级检测：vim > nvim > nano > vi
    - [x] Windows: 优先使用 VS Code、Notepad++、Sublime Text
    - [x] 如果不可用，回退到 `notepad.exe`
  - [x] 编写 `octopus config edit` 命令测试
  - [x] `octopus config edit` 命令实现
    - [x] 获取当前加载的配置文件路径
    - [x] 自动检测并启动系统默认编辑器
    - [x] 编辑后提示用户重启服务以应用更改
    - [x] 支持 `--editor` 参数指定特定编辑器
  - [x] 编写配置文件语法验证测试
  - [x] 配置文件更改后自动验证
    - [x] TOML 语法检查
    - [x] 配置字段完整性验证
    - [x] 如果配置无效，提示用户并允许重新编辑
  - [x] 编写安全性检查测试
  - [x] 编辑器安全性考虑
    - [x] 防止编辑器参数注入
    - [x] 验证编辑器可执行文件存在性
    - [x] 处理编辑器启动失败的情况

### Phase 9: 跨平台环境变量管理 (新增需求)
- [ ] 环境变量自动设置功能 (测试先行)
  - [ ] 编写跨平台环境变量设置测试
  - [ ] Linux/macOS 环境变量设置
    - [ ] 检测用户使用的 shell (bash, zsh, fish)
    - [ ] 自动向对应配置文件添加环境变量 (~/.bashrc, ~/.zshrc)
    - [ ] 支持临时设置 (当前会话) 和永久设置
  - [ ] Windows 环境变量设置
    - [ ] 使用 Windows API 设置用户环境变量
    - [ ] 支持 PowerShell 和 CMD 环境
    - [ ] 注册表操作 (HKEY_CURRENT_USER\Environment)
  - [ ] 编写 `octopus env set` 命令测试
  - [ ] `octopus env set` - 自动配置 Claude Code 环境变量
    - [ ] 设置 ANTHROPIC_BASE_URL=http://localhost:8080
    - [ ] 设置 ANTHROPIC_API_KEY=dummy-key
    - [ ] 支持 `--permanent` 参数永久设置
    - [ ] 支持 `--shell` 参数指定特定 shell
  - [ ] 编写 `octopus env unset` 命令测试
  - [ ] `octopus env unset` - 移除环境变量设置
  - [ ] 编写 `octopus env show` 命令测试
  - [ ] `octopus env show` - 显示当前环境变量状态

### Phase 10: 自动更新和CI/CD系统 (已完成自动更新功能)
- [x] 版本检测系统 (测试先行)
  - [x] 编写版本比较逻辑测试
  - [x] 实现语义化版本比较 (semver)
  - [x] 编写GitHub API版本检查测试
  - [x] GitHub Releases API集成 (VibeAny/octopus-cli)
  - [x] 编写版本缓存和检查频率测试
  - [x] 本地版本缓存机制 (避免频繁请求)
- [x] 自动更新提示功能 (测试先行)
  - [x] 编写启动时版本检查测试
  - [x] 启动时自动检查新版本 (可配置)
  - [x] 编写用户交互提示测试
  - [x] 美观的更新提示界面 (彩色输出)
  - [x] 编写更新确认逻辑测试
  - [x] 用户同意/拒绝更新处理
- [x] 自动下载更新功能 (测试先行)
  - [x] 编写平台检测测试
  - [x] 自动检测当前平台和架构
  - [x] 编写下载进度显示测试
  - [x] 下载进度条和状态显示
  - [x] 编写文件校验测试
  - [x] 下载文件完整性校验 (SHA256)
  - [x] 编写备份和替换测试
  - [x] 原程序备份和新版本替换
- [x] 手动升级命令 (测试先行)
  - [x] 编写 `octopus upgrade` 命令测试
  - [x] `octopus upgrade` - 手动检查升级
  - [x] 编写 `octopus upgrade --force` 测试
  - [x] 强制升级选项
  - [x] 编写 `octopus upgrade --check` 测试
  - [x] 仅检查不升级选项
- [ ] GitHub Actions CI/CD流水线 (VibeAny/octopus-cli)
  - [ ] 构建流水线配置
    - [ ] 多平台构建矩阵 (8个平台)
    - [ ] 自动化测试执行
    - [ ] 代码质量检查 (lint, vet, security scan)
  - [ ] 发布流水线配置
    - [ ] Tag触发自动构建
    - [ ] 语义化版本管理
    - [ ] 自动生成Release Notes
    - [ ] 多平台二进制打包和上传到GitHub Releases
    - [ ] 校验和文件生成 (SHA256SUMS)
  - [ ] 安全和权限管理
    - [ ] GitHub Secrets配置
    - [ ] GPG签名支持 (可选)
    - [ ] 发布权限控制

### Phase 11: 测试和文档
- [ ] 单元测试
  - [ ] CLI 命令测试
  - [ ] 配置管理器测试
  - [ ] 转发引擎测试
  - [ ] 进程管理器测试
- [ ] 集成测试
  - [ ] 端到端 CLI 测试
  - [ ] 代理功能测试
  - [ ] 配置切换测试
- [ ] 文档完善
  - [ ] CLI 命令文档
  - [ ] 用户使用指南
  - [ ] 配置文件说明
  - [ ] 故障排除指南

### Phase 12: 发布准备 (已完成)
- [x] 构建和打包
  - [x] 多平台构建脚本 (Makefile 支持 8 个平台)
  - [x] 二进制文件打包 (版本化命名规则)
  - [x] 安装脚本生成 (install.sh 自动安装脚本)
- [x] 发布流程
  - [x] 版本管理 (语义化版本 + Git SHA)
  - [x] 发布说明 (README 文档完整)
  - [x] 分发渠道准备 (支持 Homebrew, Chocolatey, Snap, Scoop, AUR)

## CLI 命令开发优先级

**✅ High Priority (MVP 必需) - 已完成:**
1. ✅ `octopus start/stop/status` - 基础服务管理
2. ✅ `octopus config add/list/switch` - 核心配置管理
3. ✅ 基础代理转发功能
4. ✅ 便捷启动模式 (`octopus` 无参数自动启动)
5. ✅ 智能配置管理 (自动记住配置文件状态)

**✅ Medium Priority - 已完成:**
1. ✅ `octopus config remove/show` - 完整配置管理
2. ✅ `octopus health` - 健康检查
3. ✅ `octopus logs` - 日志查看
4. ✅ 配置验证和错误处理
5. ✅ API 即时切换 (自动重启守护进程)
6. ✅ 可移植配置系统

**⏳ Low Priority - 待开发:**
1. `octopus restart` - 便利功能
2. 管理 API 接口
3. 监控和统计功能

**✅ New Priority (新增需求) - 已完成:**
1. ✅ `octopus config edit` - 配置文件编辑功能 (使用系统默认编辑器)

**🆕 New Priority (新增需求) - 待开发:**
1. `octopus env set/unset/show` - 环境变量自动设置功能
2. `octopus upgrade` - 自动升级功能
3. 版本检测和提示系统
4. GitHub Actions CI/CD流水线
5. 自动化发布系统

## 开发规范

### 开发方法论 - TDD (Test-Driven Development)
本项目采用 **测试驱动开发 (TDD)** 方法论，强调测试优先：

#### TDD 开发流程
1. **Red**: 编写失败的测试用例
2. **Green**: 编写最小可行代码使测试通过
3. **Refactor**: 重构代码保持测试通过

#### TDD 实践原则
- 任何新功能开发前必须先编写测试
- 测试用例要覆盖正常流程和边界条件
- 保持测试代码的可读性和可维护性
- 每个 commit 都应该包含对应的测试
- 持续运行测试确保回归

### CLI 设计原则
- 命令结构清晰，易于记忆
- 提供丰富的帮助信息
- 支持常用的标志参数 (--verbose, --config, --help)
- 错误信息友好且具有指导性

### 代码规范
- 使用 `gofmt` 和 `goimports` 格式化代码
- 遵循 Go 命名约定
- CLI 命令使用 Cobra 框架
- 配置管理使用标准 TOML 库
- 严格遵循 TDD 开发流程

### 配置管理规范  
- TOML 格式配置文件
- 支持配置文件优先级
- 敏感信息安全存储
- 配置变更原子操作

### 测试规范 (TDD 重点)
- **单元测试**: 每个函数/方法都要有对应测试
- **集成测试**: CLI 命令端到端测试
- **测试覆盖率**: 目标 > 90% (比之前提高)
- **测试命名**: 使用 `TestFunctionName_Scenario_ExpectedBehavior` 格式
- **表驱动测试**: 使用表格驱动测试处理多种场景
- **Mock 和 Stub**: 模拟外部依赖 (API 调用、文件系统)
- **测试金字塔**:
  - 单元测试 (70%): 快速、独立、可靠
  - 集成测试 (20%): 组件间交互
  - 端到端测试 (10%): 完整用户场景

### Git 和开源规范
- 使用语义化提交信息 (Conventional Commits)
- 功能开发使用分支 (feature/功能名)
- Pull Request 必须包含测试
- 代码审查必须通过所有测试
- MIT 开源许可证

## 技术债务管理

### 当前技术债务
- [ ] 错误处理标准化
- [ ] 日志格式统一
- [ ] 配置验证完善
- [ ] 测试覆盖率提升

### 性能优化
- [ ] HTTP 客户端连接池
- [ ] 配置热重载优化
- [ ] 内存使用优化
- [ ] 启动时间优化

## 里程碑

1. ✅ **CLI 框架版本** - 基础 CLI 命令和配置管理 (Phase 2-3) [已完成]
2. ✅ **MVP 版本** - 基本代理和配置切换功能 (Phase 4-5) [已完成]
3. ✅ **Beta 版本** - 完整 CLI 功能和用户体验优化 (Phase 6-7) [核心功能已完成]
4. ✅ **增强版本** - 配置编辑功能和用户体验提升 (Phase 8) [已完成]
5. ⏳ **环境变量版本** - 跨平台环境变量管理 (Phase 9) [待开发]
6. ✅ **自动更新版本** - 自动更新功能 (Phase 10) [已完成]
7. ⏳ **CI/CD版本** - GitHub Actions流水线 (Phase 10 剩余部分) [待开发]
8. ⏳ **正式版本** - 完善测试和文档 (Phase 11) [测试覆盖待完善]  
9. ✅ **发布版本** - 可分发安装 (Phase 12) [已完成]

## 质量门禁

### ✅ MVP 版本要求 - 已完成
- ✅ 基础 CLI 命令正常工作
- ✅ TOML 配置文件读写正常
- ✅ 代理转发功能正常
- ✅ 配置切换功能正常

### ✅ Beta 版本要求 - 已完成
- ✅ 所有 CLI 命令实现
- ✅ 错误处理完善
- ✅ 用户体验友好 (便捷启动、智能配置管理、可移植部署)
- ✅ 基本测试覆盖

### 🆕 当前版本特色功能
- ✅ **便捷启动**: `octopus` 无参数直接启动，无需指定配置文件
- ✅ **智能配置**: 自动记住当前使用的配置文件，支持无缝切换
- ✅ **即时生效**: API 切换后自动重启守护进程，立即应用新配置
- ✅ **完整日志**: 记录所有请求转发、API 切换、服务重启事件
- ✅ **可移植部署**: 所有配置文件、日志、PID 文件保存在二进制同目录
- ✅ **健壮错误处理**: 配置文件不存在时自动回退，API 不存在时提示错误
- ✅ **彩色输出**: 美观的彩色表格和状态显示，支持 `--no-color` 选项，已修复表格对齐问题
- ✅ **多平台构建**: 支持 Windows, macOS (x64/ARM64), Linux 平台构建 (8个平台)
- ✅ **版本化发布**: 规范的版本命名 `v0.0.1-platform-YYYYMMDD.git_sha`
- ✅ **配置编辑**: `octopus config edit` 使用系统默认编辑器便捷修改配置
- ✅ **跨平台编辑器支持**: 自动检测 vim/VS Code/Notepad++ 等主流编辑器
- ✅ **配置验证**: 编辑后自动验证 TOML 语法和配置完整性
- ✅ **日志跟踪**: `octopus logs -f` 实时跟踪服务日志输出
- ✅ **一键安装**: 支持 curl/wget 一键安装脚本 `install.sh`
- ✅ **包管理器支持**: 支持 Homebrew, Chocolatey, Snap, Scoop, AUR 多种分发渠道
- ✅ **完整文档**: 中英文 README，详细的使用指南和配置说明
- ✅ **自动升级**: `octopus upgrade` 命令支持一键升级到最新版本
- ✅ **版本管理**: 语义化版本比较，GitHub Releases API 集成
- ✅ **安全更新**: 文件完整性校验，自动备份和恢复机制
- ✅ **更新进度**: 实时下载进度显示和用户确认机制

### 正式版本要求
- 完整测试覆盖
- 文档完善
- 性能满足要求
- 安全审计通过