# Phase 2: TDD 多阶段修复计划

> 状态：进行中。本计划基于 2026-08-02 对当前工作树（含 phase-1 未提交修复）的重新审计。

## 背景

phase-1（2026-07-22 计划）已修复：跨卷回滚、junction 失败孤儿文件、
迁移失败自动恢复、电池警告非阻塞、Windows Update 取值、卷标、ASCII 盘符、
前端编译、冲突 UI 死代码移除等。当前 `go test`、前端 vitest/tsc/eslint 基本全绿。

## 阶段

### Stage 1: 日志测试隔离

- `internal/migrate/log_test.go` 的 `TestLogger_NewAndClose` 直接写真实
  `%APPDATA%\MoveKit\logs`，在受限环境（沙箱/CI 只读配置）必然失败。
- 修复：测试内 `t.Setenv("APPDATA", t.TempDir())`。

### Stage 2: 已迁移文件夹幂等（批量重试修复）

- 问题：`StartMigration` 遍历全部 5 个文件夹；若某文件夹已是 junction
  （`actualSource == targetPath`），`Migrate` 的 PreCheck 会返回
  `TARGET_EXISTS`，导致整批迁移/重试失败。
- TDD：先写 `TestMigrate_AlreadyMigratedSameTargetIsNoOp`（红），再在
  `Migrate` 中加 early-return（绿）。

### Stage 3: 冲突机制收敛（删除死代码）

- README 产品语义：存在冲突时阻止迁移。但 `Migrate`/`moveFilesCrossVolume`
  保留了一套不可达的 `ConflictSkip`/`ConflictOverwrite` 机制，且冲突检测在
  PreCheck 之后执行，永远无法报告冲突数量。
- TDD：先写目标路径存在同名文件时返回带数量的 `TARGET_EXISTS` 测试（红），
  再收敛代码（绿）。`StartMigration` 公开签名保持不变（Wails 绑定无法重新生成）。

### Stage 4: 前端取消语义与释放空间统计

- 取消时前端显示“迁移失败：迁移已取消...”，应改为“迁移已取消”。
- “释放空间”把本次运行前就已迁移的文件夹也算入，应只统计本次由
  `default` 变为 `migrated` 的文件夹。
- TDD：在 `App.test.tsx` 先写两个失败用例，再实现。

## 验证命令

```powershell
$env:GOCACHE = Join-Path $env:TEMP 'movekit-gocache'
go test ./...
go vet ./...
cd frontend; npm.cmd test; npx.cmd tsc --noEmit; npm.cmd run lint
```

## 审计

每个阶段完成后，由独立子 Agent（review-agent 技能）对阶段改动做只读、
缺陷优先的审查；发现的问题在本阶段修复后再进入下一阶段。
