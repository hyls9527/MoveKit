# 质量保障记录 07 — 缺陷系统修复与验证

日期：2026-08-02

## 1. 修复清单

| ID | 严重度 | 缺陷 | 修复 | 验证 |
| --- | --- | --- | --- | --- |
| D-01 | P3 | Shell Folders 键缺失时注册表写入失败（`OpenKey` 未创建键） | `WriteString`/`WriteExpandString` 打开键时增加 `CREATE_SUB_KEY`，缺失键自动创建 | 全量回归通过；代码审查确认 |
| D-02 | P3 | `saveBackup` 0644 与 `SaveBackup` 0600 权限不一致 | 统一为 0600（仅所有者可读写）；`config.json` 同步 0600 | 全量回归通过 |
| A-1 | P2 | User Shell Folders 值类型在迁移（REG_SZ）与回滚（REG_EXPAND_SZ）间反复横跳 | 迁移写 User Shell Folders 改用 `WriteExpandString`，保持约定类型 | 全量回归通过 |
| A-2 | P3 | `ReadInteger` 错误消息含双反斜杠（HKCU\\） | 修正格式化串 | vet/测试通过 |

## 2. 单元/回归验证

```
go test ./...     → 93 PASS / 0 FAIL
go vet ./...      → 0 告警
gofmt             → 全部改动文件已格式化
vitest            → 12 PASS
eslint / tsc      → 0 错误
```

## 3. 修复影响面

- D-01：仅影响注册表写入路径，读取行为不变；对正常存在 Shell Folders 键的系统无行为变化；
- D-02：文件权限收紧，不影响读写逻辑；
- A-1：仅改变写入的注册表值类型（EXPAND_SZ），路径内容不变；回滚逻辑与之一致。

## 4. 结论

全部已登记缺陷修复完毕，回归用例通过，未引入新风险。
