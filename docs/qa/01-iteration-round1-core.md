# 质量保障记录 01 — 第一轮迭代优化（核心功能）

日期：2026-08-02
基线：未提交工作区（相对 `bc1e8ce` 的多阶段修复 + 独立审计修复）

## 1. 迭代目标

聚焦核心功能模块（迁移引擎、回滚、junction 解析），保证功能完整性与逻辑正确性：

- 修复已识别的核心逻辑缺口，不引入新行为回归；
- 为关键路径补充回归测试；
- 确认技术架构稳定（分层、mock 抽象、数据流不变）。

## 2. 迭代范围

| 模块 | 范围 |
| --- | --- |
| `internal/migrate` | 盘符校验、回滚幂等性、失败恢复链路 |
| `internal/syscall` | junction reparse buffer 解析边界防护 |
| 测试 | 以上修复对应的单元/管道测试 |

## 3. 验收标准

- `go test ./...` 全绿，`go vet` 无告警；
- 新增测试覆盖：畸形 reparse buffer、小写盘符、回滚幂等重试；
- 不修改任何对外 Wails 绑定签名。

## 4. 实施内容

### 4.1 盘符校验大小写不一致（逻辑正确性）

原 `Migrate` 只接受大写 `A-Z`，而 `PreCheck` 接受小写并归一化为大写，同一目标盘传 `"d"` 时两个函数行为不一致。

修复（`internal/migrate/migrate.go`）：统一使用 `isASCIIDriveLetter` 校验并 `strings.ToUpper` 归一化。

### 4.2 回滚非幂等（异常流程）

原 `Rollback` 假设 `NewPath` 一定存在。若上一次回滚已把文件移回原处、仅注册表恢复失败，用户重试回滚会因 `NewPath` 不存在而再次失败，形成死循环。

修复（`internal/migrate/rollback.go`）：回滚前检测 `NewPath`：

- `NewPath` 存在 → 正常移动文件；
- `NewPath` 不存在且 `OriginalPath` 存在 → 跳过移动，直接完成注册表恢复（幂等重试）；
- 两者均不存在 → 返回明确错误，提示数据可能被手动删除。

### 4.3 junction reparse buffer 解析边界（安全/健壮性）

原 `ResolveTarget` 内联解析 `REPARSE_DATA_BUFFER`，无畸形数据防护测试。

修复（`internal/syscall/junction.go`）：抽出 `parseMountPointTarget`，明确校验 buffer 最小长度、reparse tag、substitute name 偏移/长度越界，避免越界读取。

## 5. 新增测试

| 测试 | 覆盖点 |
| --- | --- |
| `TestParseMountPointTarget` | 合法解析、截断 buffer、错误 tag、偏移越界 |
| `TestMigrate_AcceptsLowercaseDriveLetter` | 小写盘符归一化后注册表写入 `D:\Desktop` |
| `TestRollback_IdempotentRetryWhenFilesAlreadyRestored` | 文件已回、注册表未恢复时的幂等重试 |

## 6. 验证结果

```
go test ./...        PASS（main / migrate / syscall）
go vet ./...         PASS
gofmt                全部改动文件已格式化
```

## 7. 技术架构稳定性确认

- 分层结构未变：`main` → `internal/migrate` → `internal/syscall`，依赖方向无环；
- Wails 绑定签名未变（`StartMigration(targetDrive, conflictMode)` 等保持兼容）；
- `Sys` 可替换层、`testTargetRoot` 测试钩子、全局状态清理机制保持不变；
- 新增逻辑均落在已有层内（解析归属 syscall、业务判断归属 migrate）。

结论：第一轮迭代目标达成，核心功能逻辑正确性得到加强，架构稳定。
