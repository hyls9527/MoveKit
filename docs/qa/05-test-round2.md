# 质量保障记录 05 — 第二轮系统性测试（异常 / 边界 / 安全深化）

日期：2026-08-02

## 1. 测试范围

本轮聚焦异常流程、边界条件与安全审查，覆盖全部核心业务场景的失败路径。

## 2. 异常与边界测试

### 2.1 新增边界用例

| 用例 | 验证点 |
| --- | --- |
| `TestGetFirstRunFlag_MalformedConfig` | config.json 损坏时按首次运行处理（不崩溃） |
| `TestLoadBackup_MalformedJSON` | 备份 JSON 损坏时返回错误而非静默成功 |
| 既有：畸形 reparse buffer、回滚幂等、取消恢复、部分复制失败恢复 | 数据安全边界 |

### 2.2 失败路径矩阵（回归确认）

| 场景 | 预期 | 结果 |
| --- | --- | --- |
| 复制中途失败 | 自动恢复，源目录文件完整 | PASS |
| 注册表写入失败 | 自动恢复 + 备份标记 rolled_back | PASS |
| 用户取消 | CANCELLED + 自动恢复 | PASS |
| 回滚时注册表失败 | 文件已回原处，可重试 | PASS |
| 目标与源均为同一路径 | no-op，保留原备份 | PASS |
| 目标盘非 NTFS/可移动/网络 | 明确错误码 | PASS |
| RebootRequired=0 | 不误报 | PASS |

## 3. 安全测试（深化）

### 3.1 依赖完整性

```
go mod verify → all modules verified
```

### 3.2 危险模式静态扫描结论

| 模式 | 位置 | 评估 |
| --- | --- | --- |
| `exec.Command("cmd","/c","mklink",...)` | junction.go 回退路径 | 参数由 Go 引号化，注入面低；路径含 `&` 等字符时仍有 cmd 解析风险（已知残余风险，P3） |
| `os.RemoveAll` | 恢复/清理路径 | 目标均为本次迁移计算出的临时/目标路径，删除前有 junction 指向校验 |
| `unsafe.Pointer` | Win32 syscall 封装 | 全部集中在 `internal/syscall`，符合 Go 调用 Windows API 的必需用法 |
| 文件权限 | backup.go / migrate.go | **发现 D-02**：`saveBackup` 写 0644，`SaveBackup` 写 0600，不一致 |
| 注册表键 | registry.go | **发现 D-01**（第一轮）：写路径未用 `CREATE_SUB_KEY`，键缺失时失败 |
| 日志脱敏 | log.go maskPath | USERPROFILE 前缀被替换，无用户名泄漏 |

### 3.3 工具缺口（记录在案）

- 本机无 gcc：`-race` 竞态检测无法本地执行，CI（windows-latest + gcc）覆盖；
- 本机无 golangci-lint / govulncheck，且网络受限：SAST 与依赖 CVE 扫描未执行，列为交付风险项。

## 4. 缺陷记录（第二轮新增）

| ID | 严重度 | 描述 | 状态 |
| --- | --- | --- | --- |
| D-02 | P3 | `saveBackup` 文件权限 0644 与 `SaveBackup` 0600 不一致，备份记录应统一为仅所有者可读写 | 待修复 |

## 5. 测试结论

异常与边界路径表现稳定；安全审查未发现高危注入或越权问题；新增 2 个低危一致性/兼容性缺陷（D-01、D-02），转入缺陷修复清单。
