# 质量保障记录 08 — 全量数据真实性自检

日期：2026-08-02

## 1. 自检范围

业务数据（备份记录、迁移状态）、配置数据（config.json、前端 localStorage）、注册表数据、日志数据。

## 2. 逐项核查

### 2.1 配置数据 config.json

| 项 | 核查结果 |
| --- | --- |
| 来源 | `%APPDATA%\MoveKit\config.json`，仅存 `firstRun` 布尔 |
| 准确性 | 缺失/损坏时按首次运行处理（有测试） |
| 完整性 | 写入为原子 JSON（MarshalIndent + WriteFile），权限 0600 |
| 一致性 | 前端实际使用 localStorage 引导，config.json 为兼容保留（双机制，无冲突） |

### 2.2 备份数据 backups/*.json

| 项 | 核查结果 |
| --- | --- |
| 字段 | Version/RegValueName/OriginalPath/NewPath/RegistryValue/CreatedAt/Status |
| 状态机 | in_progress → complete / rolled_back；失败自动恢复会标记 rolled_back |
| 准确性 | 成功迁移路径写入目标绝对路径；回滚保留原始记录不被 no-op 覆盖（新增测试） |
| 完整性 | JSON 往返测试通过；损坏文件 LoadBackup 返回错误；LoadAllBackups 跳过损坏项（见 3.1） |
| 权限 | 统一 0600（D-02 修复） |

### 2.3 注册表数据

| 项 | 核查结果 |
| --- | --- |
| User Shell Folders | 迁移写入 REG_EXPAND_SZ（A-1 修复后类型不再横跳），值为 `D:\Desktop` 形式绝对路径 |
| Shell Folders | 写入 REG_SZ 绝对路径；键缺失时自动创建（D-01 修复） |
| 双键一致性 | 迁移/回滚均双键写入，任一次失败有明确错误与恢复路径 |
| 合规性 | 仅写 HKCU 当前用户，不要求管理员权限 |

### 2.4 状态数据

`state.json` 相关代码已作为死代码删除（审计阶段），系统不再产生该文件；架构文档中提及的 state.json 与实际实现不一致，记录为文档遗留（3.3）。

### 2.5 前端本地数据

`localStorage: movekit-first-run-seen`：首次引导完成后写入 `"true"`，读写一致，无敏感数据。

### 2.6 日志数据

`%APPDATA%\MoveKit\logs\movekit-{date}.log`：等级 INFO/WARN/ERROR，互斥写入；路径经 `maskPath` 脱敏，无用户名泄漏。

## 3. 异常数据项与处理结论

| ID | 异常 | 处理结论 |
| --- | --- | --- |
| A-1 | User Shell Folders 类型横跳 | 已修复（WriteExpandString），回归通过 |
| A-2 | `LoadAllBackups` 静默跳过损坏 JSON | 保留现状（读路径不崩溃）；建议后续对损坏项输出告警，非交付阻塞 |
| A-3 | `BackupRecord` 在 migrate 与 app_types 双份定义（字段略异：RegValueName vs Folder/GUID） | 功能一致（以 migrate 为准）；类型漂移记录为维护风险 |
| A-4 | OriginalPath 存展开后路径，丢失 `%USERPROFILE%` 原形 | 已知限制：回滚写入展开路径，功能正确；记录不修改 |

## 4. 数据来源可靠性

- 迁移源路径来自注册表实读（非硬编码），junction 目标经 `ResolveTarget` 实读 reparse point；
- 测试全程使用临时目录与 mock 注册表，未接触真实用户数据；
- 前端 mock 数据仅存在于 dev fallback，不持久化。

## 5. 自检结论

核心数据准确性、完整性、一致性达标；发现并修复 1 项类型一致性异常（A-1），其余为记录在案的已知限制。
