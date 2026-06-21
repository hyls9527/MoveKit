# MoveKit — C盘瘦身工具

## What This Is

MoveKit 是一个 Windows 桌面应用，帮助用户一键迁移用户文件夹（桌面/下载/文档/图片/音乐）到其他盘。通过 NTFS Junction 技术实现100%兼容迁移，所有程序照常使用，无需改变使用习惯。单文件便携 exe，双击即用。

## Core Value

**安全、一键完成 C盘用户文件夹瘦身，迁移后所有软件正常工作，可随时回滚。**

## Requirements

### Validated

- ✓ NTFS Junction 创建、检测、删除 — `internal/syscall/junction.go`
- ✓ 5 个 Windows Known Folder 统一迁移（桌面/下载/文档/图片/音乐） — `internal/folders/known.go`
- ✓ 备份记录持久化，支持一键回滚 — `internal/migrate/backup.go` + `rollback.go`
- ✓ OneDrive KFM + 第三方云同步（Google Drive/iCloud）冲突检测 — `internal/migrate/onedrive.go` + `cloud.go`
- ✓ 磁盘空间、NTFS 文件系统、BitLocker、Windows Update、电池状态预检 — `internal/migrate/check.go`
- ✓ 实时进度上报（文件数/字节数/百分比） — `internal/migrate/migrate.go` + `app.go`
- ✓ 取消正在进行的迁移 — `internal/migrate/move.go` 原子标志
- ✓ 单实例互斥锁防重复启动 — `internal/syscall/process.go`
- ✓ 结构化错误码 + 中文用户提示 — `internal/migrate/errors.go`
- ✓ 文件日志（按日轮转，路径脱敏） — `internal/migrate/log.go`
- ✓ Windows 注册表读写（HKCU User Shell Folders / Shell Folders） — `internal/syscall/registry.go`
- ✓ 便携单 exe 分发，WebView2 运行时依赖 — `main.go` embed + `wails.json`

### Active

- [x] **BUG-01**: 修复 `process.go:43` unsafe type assertion panic 风险 (Phase 1)
- [x] **BUG-02**: 修复 junction debug buffer 大小不匹配 (Phase 1)
- [x] **BUG-03**: 修复 backup save errors 被静默吞掉 (Phase 1)
- [x] **BUG-04**: 修复同盘迁移时无降级方案（应直接用 os.Rename） (Phase 1)
- [x] **BUG-05**: 修复 `dirSize` 错误被静默忽略导致进度不准 (Phase 1)
- [x] **BUG-06**: 修复 shell notify UTF16 编码错误被忽略 (Phase 1)
- [x] **BUG-07**: 修复 `move.go:137` percent 死代码 (Phase 1)
- [x] **BUG-08**: 修复安全软件检测结果被丢弃 (Phase 1)
- [ ] **TEST-01**: migrate 包全部源文件有对应测试（当前 7/17 文件有测试; +2 in Phase 1）
- [ ] **TEST-02**: syscall 包全部源文件有对应测试（当前 8/7 文件有测试; +2 in Phase 1 — COMPLETE)
- [ ] **TEST-03**: app.go Wails 绑定方法有测试覆盖（当前 1 test file; +1 in Phase 1）
- [ ] **TEST-04**: 前端关键组件有基础测试（当前零测试）
- [x] **IMPROVE**: IMP-01 error audit, IMP-02 real progress, IMP-03 cancel wiring, IMP-04 ResolveConflict error (Phase 1)

### Out of Scope

- 新文件夹类型扩展 — 当前仅支持 5 个 Known Folder，保持范围不变
- UI 大改版 — 本轮聚焦质量，视觉整体翻新另起 phase
- 安装包/签名 — 保持便携 exe 分发方式
- 国际化/多语言 — 保持中文
- macOS/Linux 支持 — Windows only

## Context

- **代码库**: `C:/Users/Admin/Desktop/Vibe coding/MoveKit`, 仓库 `github.com/hyls9527/MoveKit`
- **版本**: 0.1.0，已可构建运行
- **技术栈**: Go 1.23.5 + Wails v2.12 + React 19 + TypeScript 6 + Vite 8 + TailwindCSS v4
- **代码库映射**: `.planning/codebase/` 下 7 份分析文档已完成
- **已知问题**: 8 个 bug（详见 CONCERNS.md），18/26 Go 源文件无测试，前端零测试
- **用户反馈**: 无（尚未发布收集）

## Constraints

- **平台**: Windows 10/11 only（Win32 API 深度绑定）
- **目标盘**: NTFS only（Junction 依赖 NTFS 重解析点）
- **权限**: 用户级 HKCU 注册表写入，不要求管理员权限
- **兼容性**: 保持单 exe 便携分发，无安装器
- **语言**: Go 后端 + TypeScript 前端，不改技术栈

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 本轮聚焦质量（bug + 测试），不扩功能范围 | 用户明确：质量为主，小改进顺手 | — Pending |
| 测试覆盖率目标 80%+，覆盖全部无测试文件 | 当前 ~30%，核心路径无保障 | — Pending |
| Wails v2 + React 不变 | 现有架构工作正常，无迁移必要 | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-06-21 after initialization*
