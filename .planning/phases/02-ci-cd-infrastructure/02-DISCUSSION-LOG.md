# Phase 2: CI/CD Infrastructure - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-06-22
**Phase:** 2-ci-cd-infrastructure
**Areas discussed:** 触发策略, Lint策略, 覆盖门禁, 构建产物, CI缓存, 前端Lint, 路径过滤, Changelog, Workflow结构, 并发控制, 版本锁定, 状态徽章, 超时, Race检测, Dependabot, 测试报告

---

## 触发策略

| Option | Description | Selected |
|--------|-------------|----------|
| Push + PR 双触发 | 每次 push 到任意分支都跑 CI，PR 也跑 | |
| 仅 PR 触发 | 只在 PR 打开/更新时跑 CI | |
| Push 触发 + main 分支保护 | push 即触发，main 要求 PR + CI 通过才能合并 | ✓ |

**User's choice:** Push 触发 + main 分支保护

---

## Lint 策略

| Option | Description | Selected |
|--------|-------------|----------|
| 默认配置 + 阻塞 PR | golangci-lint 默认 linter 集，阻塞合并 | |
| 宽松模式（仅报告） | 运行 lint 但失败不阻塞 | |
| 严格模式（自定义规则集） | 自定义规则，零警告阻塞 | ✓ |

**User's choice:** 严格模式 — 自定义 `.golangci.yml`，启用全面规则

---

## 覆盖门禁

| Option | Description | Selected |
|--------|-------------|----------|
| 仅生成报告，不设门禁 | CI 运行覆盖率但不阻塞 | |
| 软门禁（当前值不降级） | 基线 ~56%，不允许降级 | |
| 直接设 80% 硬门禁 | 立即 80% 门禁 | |
| 分阶段：先报告后阻塞 | Phase 2 仅报告，Phase 3 升级为阻塞 | ✓ |

**User's choice:** 分阶段 — Phase 2 仅报告不阻塞，Phase 3 完成后升级

---

## 构建产物

| Option | Description | Selected |
|--------|-------------|----------|
| 仅验证构建通过 | 只运行 wails build，不保留产物 | |
| 上传 exe 为 CI artifact | 保留 exe 7-30 天 | |
| 构建 + 上传到 Release | tag push 时自动创建 Release + 上传 exe | ✓ |

**User's choice:** 构建 + 上传到 Release

---

## CI 缓存策略

| Option | Description | Selected |
|--------|-------------|----------|
| 全部缓存 | Go build + modules + npm + lint cache | ✓ |
| 仅 Go + npm | 不缓存 build 和 lint cache | |
| 仅 Go modules | 最少维护负担 | |

**User's choice:** 全量缓存，hash-based key

---

## 前端 Lint 集成

| Option | Description | Selected |
|--------|-------------|----------|
| ESLint + tsc --noEmit | 两层检查，与 Go lint 同级阻塞 | ✓ |
| 仅 tsc --noEmit | 只检查类型，ESLint 留给 Phase 4 | |
| 跳过前端 Lint | Phase 2 只管 Go | |

**User's choice:** ESLint + tsc --noEmit

---

## 路径过滤策略

| Option | Description | Selected |
|--------|-------------|----------|
| 始终全量运行 | 不跳过任何变更 | ✓ |
| 跳过纯文档变更 | .md/.planning 变更时跳过 | |
| 按目录拆分运行 | Go 只跑 Go，前端只跑前端 | |

**User's choice:** 始终全量运行，无 [skip ci]

---

## Changelog 自动生成

| Option | Description | Selected |
|--------|-------------|----------|
| 基于 Conventional Commits | 工具自动生成 changelog | ✓ |
| 手动编写 Release Notes | 人在 GitHub Release 编辑器写 | |

**User's choice:** Conventional Commits 自动生成

---

## Workflow 文件拆分

| Option | Description | Selected |
|--------|-------------|----------|
| 单一 ci.yml | 一个文件包含所有 job | ✓ |
| 拆分为 ci.yml + release.yml | CI 和 Release 分开 | |

**User's choice:** 单一 ci.yml

---

## 并发控制

| Option | Description | Selected |
|--------|-------------|----------|
| 取消旧运行 | 同一分支新 push 取消旧 CI | ✓ |
| 允许并发 | 不取消，每次独立完成 | |

**User's choice:** cancel-in-progress

---

## 工具版本锁定

| Option | Description | Selected |
|--------|-------------|----------|
| 全部锁定 | Go/Node/golangci-lint 版本全锁 | ✓ |
| 仅锁 Go 和 Node | golangci-lint 用 latest | |
| 全部 latest | 全部用最新 | |

**User's choice:** 全部锁定

---

## 状态徽章

| Option | Description | Selected |
|--------|-------------|----------|
| 添加到 README | CI 通过后展示 badge | ✓ |
| 不需要 | 保持 README 简洁 | |

**User's choice:** 添加到 README

---

## 超时设置

| Option | Description | Selected |
|--------|-------------|----------|
| 15 分钟 | 给足余量同时防止挂死 | ✓ |
| 默认 360 分钟 | GitHub Actions 默认 | |

**User's choice:** 15 分钟

---

## Race 检测

| Option | Description | Selected |
|--------|-------------|----------|
| 启用 -race | windows-latest 有 MinGW 可用 | ✓ |
| 跳过 -race | 减少 CI 时间约 30% | |

**User's choice:** 启用 -race

---

## Dependabot

| Option | Description | Selected |
|--------|-------------|----------|
| 仅 Go 依赖 | Go modules 自动更新 | |
| Go + npm 都启用 | 两端都自动更新 | ✓ |
| 暂不启用 | 手动管理依赖 | |

**User's choice:** Go + npm 都启用

---

## 测试报告

| Option | Description | Selected |
|--------|-------------|----------|
| 上传 JUnit 报告 | 结构化 XML，PR 摘要可查看 | ✓ |
| 仅控制台输出 | 不看结构化报告 | |

**User's choice:** JUnit XML artifact

---

## Release 触发细节

| Option | Description | Selected |
|--------|-------------|----------|
| 仅语义化标签 (v*) | 仅 v0.2.0 风格标签触发 | ✓ |
| 任意 tag | 任何 tag 都触发 | |
| 手动触发 | workflow_dispatch | |

**User's choice:** 仅 v* 标签触发

## 版本来源

| Option | Description | Selected |
|--------|-------------|----------|
| Git tag 驱动 | 提取 tag 版本号，ldflags 注入 | ✓ |
| 手动指定 | workflow_dispatch input | |

**User's choice:** Git tag 驱动 + ldflags 注入

## Release 模式

| Option | Description | Selected |
|--------|-------------|----------|
| 自动发布 | tag push 直接创建正式 Release | ✓ |
| 始终 Draft | draft 由人审核后发布 | |

**User's choice:** 自动发布

## Wails CLI 缓存

| Option | Description | Selected |
|--------|-------------|----------|
| 缓存二进制 | 基于 Wails 版本号的缓存 | ✓ |
| 每次重新安装 | 每次 CI 重新编译 | |

**User's choice:** 缓存二进制

## 缓存 key 策略

| Option | Description | Selected |
|--------|-------------|----------|
| 基于 hash | go.sum/package.json hash 驱动失效 | ✓ |
| 每周失效 + hash | hash 基础上叠加时间戳 | |

**User's choice:** 基于 hash

---

## Claude's Discretion

- golangci-lint `.golangci.yml` 具体 linter 选择
- ESLint 配置（规则集、parser options）
- JUnit 转换工具选择（`gotestsum` vs `go-junit-report`）
- 缓存 key 精确构造细节
- Dependabot 更新频率

## Deferred Ideas

- 多 Windows 版本矩阵测试（windows-2019 / windows-2022）
- 前端测试 CI 集成（Vitest + RTL — Phase 4）
- 代码签名
- CI 失败通知（Slack/Discord）
- 自动合并 Dependabot PR
