# 质量保障记录 02 — 第二轮迭代优化（用户体验）

日期：2026-08-02

## 1. 迭代目标

提升交互流畅度、界面友好性与操作一致性，重点覆盖：错误信息可操作性、目标盘选择引导、对话框键盘可达性。

## 2. 迭代范围

- `frontend/src/App.tsx`：迁移失败时的错误提示内容；
- `frontend/src/components/DiskSelector.tsx`：禁用盘原因说明；
- `frontend/src/components/OneDriveWarning.tsx`：Esc 关闭；
- 对应前端测试。

## 3. 验收标准

- 前端测试全绿（vitest）、ESLint 0 错误、`tsc --noEmit` 通过；
- 失败提示携带后端 `Action` 修复建议；
- 被禁用的 C 盘有明确原因说明；
- 警告对话框支持 Esc 关闭且不影响其他交互。

## 4. 实施内容

### 4.1 失败提示携带修复建议

后端错误对象包含 `code/detail/action`，但原前端只显示 `message`（如"迁移失败：无法读取目标盘卷信息"），用户不知道下一步做什么。

修复：失败事件含 `action` 时拼接显示：`迁移失败：<错误>（<建议>）`。

### 4.2 源盘禁用原因说明

DiskSelector 已禁用 C:，但未说明原因，用户可能困惑。

修复：C: 行追加"（源盘不可作为迁移目标）"提示，并保持与可移动/网络盘一致的视觉层级。

### 4.3 对话框键盘可达性

OneDriveWarning 为模态阻断型对话框，但无法用 Esc 关闭。

修复：监听 `Escape` 键调用 `onClose`，并在卸载时移除监听。

## 5. 新增/更新测试

| 测试 | 覆盖点 |
| --- | --- |
| `App.test.tsx` 失败用例 | 错误信息拼接 `action`（迁移失败：boom（请重试）） |
| `DiskSelector.test.tsx` | C: 禁用且显示"源盘不可作为迁移目标" |

## 6. 验证结果

```
vitest：12/12 通过
ESLint：0 错误
tsc --noEmit：通过
```

## 7. 架构稳定性确认

- 未改变状态机流转（idle/checking/confirming/migrating/completed/failed）；
- 未改变 Wails 绑定与数据模型；
- 新增 Esc 监听严格在组件生命周期内。

结论：第二轮迭代目标达成，交互一致性与可操作性提升，架构稳定。
