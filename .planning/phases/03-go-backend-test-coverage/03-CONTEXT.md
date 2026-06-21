# Phase 3: Go Backend Test Coverage - Context

**Gathered:** 2026-06-22
**Status:** Ready for planning
**Mode:** Autonomous (decisions derived from architecture audit + requirements)

<domain>
## Phase Boundary

Go 后端测试覆盖率达到 >=80%（migrate 包 >=85%）。为 syscall 层创建接口抽象（RegistryAccessor / ProcessChecker / VolumeChecker），使 migrate 层可脱离真实 Windows API 测试。零测试触碰真实 HKCU 注册表或泄漏 NTFS junction。
</domain>

<decisions>
## Implementation Decisions

### Syscall 接口抽象
- **D-01:** 创建 `internal/syscall/interfaces.go` — 定义 `RegistryAccessor`（ReadExpandString/WriteExpandString/WriteString/ReadInteger）、`ProcessChecker`（IsProcessRunning）、`VolumeChecker`（GetVolumeInfo/GetDiskFreeSpace/IsCrossVolume）接口
- **D-02:** Mock 框架 — `go.uber.org/mock`（官方维护的 gomock fork），通过 `//go:generate mockgen` 自动生成
- **D-03:** 现有 syscall 函数签名不变——添加接口包装层，不修改已有调用方

### 测试优先级（按风险排序）
- **D-04:** 第一批（HIGH）— `Migrate()` 集成测试、`Rollback()` 集成测试、`moveFilesCrossVolume()` 单元测试
- **D-05:** 第二批（MEDIUM）— `PreCheck()` 全路径测试、`CheckJunctionHijack()` 安全测试、`DetectOneDriveKFM()` 注册表 mock 测试
- **D-06:** 第三批（LOW）— `cloud.go`、`security.go`、`verify.go`、`state.go`、`log.go`、`version.go`

### 测试隔离
- **D-07:** 所有测试使用 `t.TempDir()` + mock 接口，零真实 Win32 API 调用
- **D-08:** 注册表测试使用 `golang.org/x/sys/windows/registry` 的 mock，不触碰真实 HKCU

### 覆盖率门禁
- **D-09:** CI 中 `go test -race -cover` 已就绪（Phase 2），覆盖率仅报告不阻塞
- **D-10:** Phase 3 完成后手动将覆盖率门禁升级为阻塞（Phase 2 的 D-07 决策）

### 架构审计发现需修复的测试 Bug
- **D-11:** `process_test.go` — mutex 名反斜杠已修复（Phase 2 顺手修）
- **D-12:** `junction_test.go` — NTFS 检查已添加（Phase 2 顺手修）
- **D-13:** `TestCreateJunction` reparse buffer 在 GitHub runner D: 盘上失败——需深入调查 BUG-02 修复是否完整

### Claude's Discretion
- 接口定义的具体方法签名
- mock 文件存放位置（`internal/syscall/mocks/` 或各包内）
- 测试辅助函数（test helpers）
- migrate.go 中 `resolveJunctionTarget` 和 `backupFilePath` 等辅助函数是否需提取为可测试接口
- app.go Wails 绑定方法的测试策略（是否需 mock Wails runtime）
</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project Definition
- `.planning/PROJECT.md` — 核心价值、约束
- `.planning/REQUIREMENTS.md` — GO-01 到 GO-05（5 个测试需求）
- `.planning/ROADMAP.md` — Phase 3 成功标准

### Architecture & Audit
- `.planning/codebase/ARCHITECTURE.md` — 依赖 DAG、类型重复、反模式
- `.planning/codebase/CONCERNS.md` — 测试覆盖缺口详细列表
- `.planning/audit/BACKEND-ARCHITECTURE-AUDIT.md` — 7.1 节：测试覆盖验证（14/25 文件有测试，56% 覆盖）

### Prior Phases
- `.planning/phases/01-bug-fix-foundation/01-CONTEXT.md` — Bug 修复决策（测试模式）
- `.planning/phases/02-ci-cd-infrastructure/02-CONTEXT.md` — CI 管线、增量 lint 策略
- `.planning/phases/02-ci-cd-infrastructure/02-RESEARCH.md` — 技术研究

### Existing Test Files
- `internal/syscall/junction_test.go` — Junction roundtrip 测试
- `internal/syscall/registry_test.go` — 注册表读写测试
- `internal/syscall/fileops_test.go` — 文件操作测试
- `internal/syscall/volume_test.go` — 磁盘信息测试
- `internal/syscall/power_test.go` — 电源状态测试
- `internal/syscall/process_test.go` — 进程枚举测试
- `internal/syscall/shell_test.go` — Shell 通知测试
- `internal/migrate/check_test.go` — PreCheck 测试
- `internal/migrate/backup_test.go` — 备份 CRUD 测试
- `internal/migrate/conflict_test.go` — 冲突检测测试
- `internal/migrate/move_test.go` — 移动/取消测试
- `internal/migrate/onedrive_test.go` — OneDrive 检测测试
- `internal/migrate/migrate_test.go` — 迁移管线测试
- `app_test.go` — dirSize 测试
</canonical_refs>

<code_context>
## Existing Code Insights

### Current Test Coverage

| Package | Files | Tested | Coverage |
|---------|-------|--------|----------|
| `internal/syscall/` | 7 | 7/7 | 100% file coverage |
| `internal/migrate/` | 15 | 6/15 | 40% file coverage |
| `app.go` + `main.go` | 3 | 1/3 | 33% (dirSize only) |
| **Total** | **25** | **14/25** | **56%** |

### Untested Critical Files (HIGH risk)
- `internal/migrate/migrate.go` — 核心迁移管线（~230 行，9 步）
- `internal/migrate/rollback.go` — 回滚逻辑
- `internal/migrate/move.go` — 文件移动（仅测试了取消标志）
- `internal/migrate/security.go` — Junction hijack 检测

### Established Test Patterns
- Go standard `testing` package（无 testify）
- `t.TempDir()` 隔离
- 测试文件与源文件同目录
- `t.Cleanup()` 清理 junction

### Integration Points
- `internal/syscall/*.go` 是 migrate 层的唯一外部依赖——抽象这层即可实现 migrate 层纯 mock 测试
- `app.go` 的 Wails-bound 方法依赖 `runtime.EventsEmit` — 测试需要 mock Wails runtime 或提取接口
</code_context>

<specifics>
## Specific Ideas

- 接口抽象优先于代码重构——先建接口层，不改变现有调用
- migrate 层测试不依赖真实 registry、真实 junction、真实进程枚举
- 覆盖率目标：80% 整体，85% migrate 包
</specifics>

<deferred>
## Deferred Ideas

- app.go Wails 绑定方法的完整测试覆盖（需 mock Wails runtime，复杂度高）— 可延后到 Phase 4 或保留为 Claude's Discretion
- 前端测试 — Phase 4 负责
- Performance benchmarks — 非本阶段范围
</deferred>

---
*Phase: 3-Go Backend Test Coverage*
*Context gathered: 2026-06-22 (autonomous mode)*
