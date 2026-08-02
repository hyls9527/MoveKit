# 质量保障记录 06 — 第三轮系统性测试（性能 / 兼容性矩阵）

日期：2026-08-02

## 1. 测试范围

性能基准（含内存分配）、大目录迁移管道耗时、资源利用率、兼容性矩阵。

## 2. 性能测试

### 2.1 基准（本机：i5-12450H / Windows 11）

```
BenchmarkDirStats-12               12     21.8 ms/op（500 文件，512B/文件）
BenchmarkParseMountPointTarget-12  3.5M   76 ns/op
```

### 2.2 迁移管道大目录测试

`BenchmarkMigratePipelineCrossVolume`：1,000 个 4KB 文件跨卷迁移（真实文件复制 + junction 创建 + 注册表模拟）：

```
1,681 ms/op（约 2.4 MB/s 有效数据速率，含统计、复制、junction、备份写盘）
3.68 MB 分配 / 30,549 次分配（约 30 次分配/文件，来自进度回调与日志）
```

### 2.3 资源利用率结论

- 进度事件节流 150ms：10 万文件场景事件量从 10 万降至数百；
- 冲突扫描跳过：目标不存在时省去一次全量目录遍历；
- 目录统计缓存：`GetFolders`/`EstimateDuration` 重复调用不重复遍历；
- 内存：`statsCache` 仅缓存 5 个已知文件夹路径，无增长风险；
- 迁移为 IO 密集型串行复制，CPU 占用低，磁盘吞吐为主导瓶颈。

## 3. 兼容性矩阵

| 维度 | 支持 | 验证方式 |
| --- | --- | --- |
| OS | Windows 10/11 | 项目约束；Windows 11 实测 |
| 文件系统 | NTFS（junction 依赖） | PreCheck 强制校验，非 NTFS 拒绝 |
| 架构 | amd64 | 本机构建验证 |
| Go | go.mod 1.23.5；本地 1.26.4 编译通过 | 双版本兼容（CI 1.23.5） |
| Node/Vite | Node 22+；Vite 8 | 本地 Node 24 构建通过 |
| WebView2 | 可 embed 打包 | CI `wails build -webview2 embed` 覆盖 |
| 注册表 | HKCU User Shell Folders / Shell Folders | 实测 User Shell Folders 可读；Shell Folders 键缺失场景记录为 D-01 |

## 4. 兼容性风险

- Shell Folders 键缺失时注册表写入失败（D-01）；
- 网络/可移动盘、非 NTFS 均被 PreCheck 拒绝（有意为之）；
- 深路径（>260 字符）：复制走 `CopyFileExW` 已加 `\\?\` 前缀（ToLongPath），Go 侧 os 调用自动处理长路径。

## 5. 测试结论

性能基线健康，无阻塞性瓶颈；兼容性矩阵与产品约束一致，缺陷集中于注册表键缺失场景（D-01）。
