# Phase 2: CI/CD Infrastructure - Context

**Gathered:** 2026-06-22
**Status:** Ready for planning

<domain>
## Phase Boundary

为 MoveKit 建立自动化 CI 管道：每次代码变更自动触发 lint（Go + TypeScript）、测试（带 race 检测 + 覆盖率报告）、构建验证（wails build）。语义化标签推送时自动创建 GitHub Release 并上传构建产物。项目当前零 CI 配置，从零搭建。
</domain>

<decisions>
## Implementation Decisions

### 触发策略
- **D-01:** Push 触发 + main 分支保护 — 每次 push 到任意分支都运行 CI，main 分支要求 PR + CI 通过才能合并
- **D-02:** 并发控制 — 同一分支/PR 有新 push 时自动取消之前未完成的 CI 运行（cancel-in-progress）
- **D-03:** 路径过滤 — 始终全量运行，不基于变更文件跳过 CI。不支持 [skip ci] 手动跳过

### Lint 策略
- **D-04:** Go Lint — 严格模式，自定义 `.golangci.yml` 配置，启用 errcheck/govet/staticcheck/gosec 等全面规则，零警告阻塞合并
- **D-05:** 前端 Lint — ESLint + `tsc --noEmit`，与 Go lint 同级阻塞
- **D-06:** 工具版本锁定 — Go 版本、Node 版本、golangci-lint 版本全部显式锁定，避免自动升级导致 CI 意外失败

### 测试与覆盖
- **D-07:** 覆盖门禁 — 分阶段实施。Phase 2 仅运行 `go test -race -cover` 生成覆盖率报告，不阻塞 CI。Phase 3（Go Backend Test Coverage）完成后将同一检查升级为阻塞门禁
- **D-08:** Race 检测 — 启用 `-race` 标志（windows-latest 已预装 MinGW，race detector 可用）
- **D-09:** 测试报告 — 输出 JUnit XML 格式，作为 CI artifact 上传，PR 摘要中可查看

### 构建与 Release
- **D-10:** Release 触发 — 仅推送语义化标签（`v*` 如 `v0.2.0`）时触发构建和发布
- **D-11:** 版本来源 — 从 git tag 提取版本号，通过 Go ldflags 注入覆盖硬编码常量
- **D-12:** Release 模式 — 自动发布（非 draft），tag push 时直接创建正式 GitHub Release 并上传 `MoveKit.exe`
- **D-13:** Changelog — 基于 Conventional Commits 规范自动生成 release notes

### CI 性能
- **D-14:** 缓存策略 — 全量缓存（Go build cache + Go module cache + npm cache + golangci-lint cache），cache key 基于 `go.sum` 和 `package.json` 的 hash
- **D-15:** Wails CLI — 缓存编译后的 wails.exe 二进制，key 基于 Wails 版本号，失败时 fallback 重新安装
- **D-16:** 超时 — 每个 job 15 分钟超时，防止挂死消耗 Runner 时长

### Workflow 结构
- **D-17:** 文件组织 — 单一 `ci.yml` 文件，包含 lint → test → build → release（条件触发）job 流水线
- **D-18:** 状态徽章 — CI 通过后 README.md 顶部展示 build status badge

### 依赖管理
- **D-19:** Dependabot — 同时启用 Go modules 和 npm 依赖自动更新，定期提交更新 PR

### Claude's Discretion
- golangci-lint `.golangci.yml` 具体 linter 选择由 Claude 根据 Go 社区最佳实践决定
- ESLint 配置（规则集、parser options）由 Claude 根据 TypeScript + React 项目标准决定
- JUnit 转换工具选择（`gotestsum` vs `go-junit-report`）由 Claude 根据生态成熟度决定
- 缓存 key 的精确构造细节由 Claude 根据 GitHub Actions cache 最佳实践决定
- Dependabot 更新频率（daily/weekly）由 Claude 根据项目节奏决定
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Definition
- `.planning/PROJECT.md` — 项目核心价值、约束、关键决策
- `.planning/REQUIREMENTS.md` — Phase 2 的 4 个 CI 需求定义（CI-01 至 CI-04）
- `.planning/ROADMAP.md` — Phase 2 目标、依赖关系、成功标准

### Codebase
- `.planning/codebase/STACK.md` — 完整技术栈（Go 1.23.5, Node 18+, Wails v2.12, React 19, Vite 8）
- `.planning/codebase/INTEGRATIONS.md` — Win32 API 依赖清单、WebView2 运行时需求、构建流程
- `.planning/codebase/ARCHITECTURE.md` — 架构约束（Windows-only, 无循环导入, 类型重复等）
- `.planning/codebase/CONCERNS.md` — 已知 Bug 列表、测试覆盖缺口（CI 需要验证所有测试通过）

### Build Configuration
- `go.mod` / `go.sum` — Go 模块定义与依赖锁定
- `frontend/package.json` — Node 依赖与构建脚本（`tsc && vite build`）
- `wails.json` — Wails 项目配置
- `frontend/tsconfig.json` — TypeScript 配置（ES2020, strict mode, react-jsx）
- `frontend/vite.config.ts` — Vite 配置（React + Tailwind 插件）

### Prior Phase Context
- `.planning/phases/01-bug-fix-foundation/01-CONTEXT.md` — Phase 1 决策（测试模式、错误处理策略、提交粒度）
</canonical_refs>

<code_context>
## Existing Code Insights

### Current State
- **零 CI 配置** — 无 `.github/workflows/` 目录，无 CI 配置文件
- **零 Lint 配置** — 无 `.golangci.yml`、`.eslintrc` 或等效配置
- **测试分布** — syscall 层 7/7 文件有测试（100%），migrate 层 6/15 文件有测试（40%），app 层 1/3 文件有测试
- **当前覆盖率** — 约 56%（14/25 Go 源文件有测试）

### Build Pipeline (manual)
1. `cd frontend && npx vite build` — 构建前端资源到 `frontend/dist/`
2. `go build -o MoveKit.exe .` — 编译 Go 二进制（embed 前端资源）
3. 或使用 `wails build` — 完整 Wails 生产构建

### Platform Requirements for CI
- **OS:** Windows only（Win32 API 深度绑定 — kernel32.dll / advapi32.dll / shell32.dll）
- **Runtime:** WebView2（windows-latest GitHub runner 预装）
- **Go:** 1.23.5（`go.mod` 指定）
- **Node:** 18+（`package.json` engines 未显式声明）
- **Wails CLI:** v2（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）

### Key Integration Points
- CI 必须能运行 `go test ./... -race`（需要 C 工具链，windows-latest 有 MinGW）
- CI 必须能运行 `wails build`（需要 WebView2 Runtime + Node.js）
- Release job 需要 `contents: write` 权限创建 Release 和上传 artifact
</code_context>

<specifics>
## Specific Ideas

无特定的视觉或行为参考——这是基础设施阶段。遵循 GitHub Actions 标准实践和 Go/Node 生态的 CI 惯例即可。
</specifics>

<deferred>
## Deferred Ideas

- **多 Windows 版本矩阵测试** — windows-2019 / windows-2022 兼容性验证。当前 windows-latest 足够，属于未来可靠性增强
- **前端测试 CI 集成** — Vitest + React Testing Library。Phase 4（Frontend Quality）建设后集成到 CI
- **代码签名** — exe 签名（需要证书）。超出当前便携分发模式范围
- **Slack/Discord 通知** — CI 失败通知。当前无通知基础设施需求
- **自动合并 Dependabot PR** — CI 通过时自动合并补丁更新。可在后续引入

None — discussion stayed within phase scope
</deferred>

---

*Phase: 2-CI/CD Infrastructure*
*Context gathered: 2026-06-22*
