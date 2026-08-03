# 第二阶段质量加固记录

日期：2026-08-02
范围：工程基础加固（PR #20，commit `17d5008`）

## 完成项

| 项目 | 结果 |
| --- | --- |
| module path 对齐 | `github.com/movekit/movekit` -> `github.com/hyls9527/MoveKit`（go.mod + 28 处引用） |
| BackupRecord 去重 | 删除 app 层死定义，migrate 包为唯一来源 |
| 损坏备份告警 | `LoadAllBackups` 返回 warnings，不再静默跳过；暴露为 `backup_corrupt` 完整性检查项 |
| 覆盖率 | main 18.5% -> 60.7%；migrate 66.5% -> 73.0%；业务代码总计 62.1% |
| 依赖安全 | golang.org/x/sys v0.30.0 -> v0.47.0（GO-2026-5024）；Go 指令 1.23.5 -> 1.25 |
| CI 加固 | 新增 govulncheck 作业；覆盖率门槛（main>=55%、business>=55%）；Go 1.25.12 |

## 期间修复的 CI 问题

- `go:embed` 依赖的 frontend/dist 移出仓库后编译失败 -> 提交 `.gitkeep` 占位
- Go 1.25.0 本身含 crypto/tls 漏洞 -> CI 钉到 1.25.12

## 结论

本阶段问题全部闭环，无遗留风险项。
