# 质量保障记录 04 — 第一轮系统性测试（功能 / 集成 / 性能 / 安全 / 兼容）

日期：2026-08-02
执行环境：Windows 11（沙箱）、Go 1.26.4 amd64、Node 24、Vite 8

## 1. 测试范围

全部核心业务场景、异常流程与边界条件，覆盖五类测试维度。

## 2. 功能测试

### 2.1 执行与结果

```
go test ./... -v   →  91 个用例 PASS，0 FAIL
vitest             →  12 个用例 PASS
go vet ./...       →  0 告警
```

### 2.2 覆盖的功能矩阵

| 模块 | 覆盖场景 |
| --- | --- |
| 迁移引擎 | 成功迁移（跨卷）、注册表失败自动恢复、部分复制失败恢复、取消恢复、电池警告放行、no-op 幂等、小写盘符、无效盘符 |
| 回滚 | 成功回滚、注册表失败、文件已回注册表未回的幂等重试、备份记录缺失 |
| junction | 创建/删除/识别、ResolveTarget、畸形 reparse buffer、非 junction 路径 |
| 检查项 | NTFS、空间、RebootRequired 值判断、ASCII 盘符、目录统计 |
| 数据持久化 | 备份 JSON 往返、LoadAllBackups、删除、配置读写 |
| 前端 | 状态机成功/失败/取消、DiskSelector 禁用与推荐、FolderList 渲染、类型定义 |

### 2.3 覆盖率

| 包 | 覆盖率 |
| --- | --- |
| package main | 15.3%（受 Wails runtime 依赖限制） |
| internal/migrate | 66.4% |
| internal/syscall | 63.1% |

## 3. 集成测试

### 3.1 后端运行时验证（cmd/verify，真实 Windows API）

15 项中 14 项通过：

✅ 磁盘信息（D: NTFS/固定盘/空间）、OneDrive KFM 检测、跨卷判断、长路径转换、冲突检测、云同步检测、备份加载、版本、日志写入、电源状态、junction 创建/删除往返、User Shell Folders 读取

❌ Shell Folders 读取：`open key HKCU\...\Explorer\Shell Folders: The system cannot find the file specified`

→ 记录为缺陷 **D-01**（见 §7）：系统缺少 Shell Folders 键时，读写直接失败；写路径未使用 `CREATE_SUB_KEY`。

### 3.2 构建集成

```
frontend: npx vite build（临时 outDir）  → 成功，371ms，28 模块
backend:  go build -o MoveKit.exe .      → 成功，6.5 MB（含 embed frontend/dist）
```

## 4. 性能测试（第一轮基线）

| 基准 | 结果 | 说明 |
| --- | --- | --- |
| `BenchmarkDirStats`（500 文件/20 目录） | 21.8 ms/op | 迁移/界面大小统计的基础操作 |
| `BenchmarkParseMountPointTarget` | 76 ns/op | junction 目标解析 |
| Vite 生产构建 | 371 ms | 前端构建耗时 |

性能风险点（第三轮深入）：大目录迁移为 IO 密集串行复制，进度事件已节流至 150ms；暂无阻塞性问题。

## 5. 安全测试（第一轮）

- 静态检查：`go vet` 0 告警；`gofmt` 全部通过；
- 代码审查要点（本轮已复核）：路径拼接使用 `filepath.Join`；注册表仅 HKCU；日志路径脱敏 `maskPath`；单实例互斥；取消为原子标志；
- 工具缺口（记录为风险）：本机无 gcc（race 检测依赖 CI）、无 golangci-lint、无 govulncheck（依赖漏洞扫描未执行，见第二轮）。

## 6. 兼容性测试（第一轮）

| 项 | 结果 |
| --- | --- |
| OS | Windows 10/11（项目约束；本机 Windows 11 验证通过） |
| 文件系统 | NTFS（PreCheck 强制校验） |
| Go 工具链 | 1.26.4 本地通过；CI 固定 1.23.5（go.mod 声明 1.23.5） |
| Node 工具链 | 24 本地通过；CI 固定 22 |
| WebView2 | 构建阶段以 embed 方式打包（`wails build -webview2 embed`，CI 覆盖） |

## 7. 缺陷记录（第一轮）

| ID | 严重度 | 描述 | 复现 | 状态 |
| --- | --- | --- | --- | --- |
| D-01 | P3 | `WriteString`/`WriteExpandString` 打开注册表键未使用 `CREATE_SUB_KEY`；Shell Folders 键缺失时迁移注册表写入失败 | cmd/verify 第 15 项 | 待修复 |

## 8. 风险与测试结论

- 主要残余风险：race 检测与 golangci-lint 未在本机执行（依赖 CI）；依赖漏洞扫描未执行（网络受限）；
- 结论：核心功能、集成链路、构建链路在第一轮测试中表现稳定，发现 1 个低危兼容性缺陷，转入缺陷修复清单。
