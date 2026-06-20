# Requirements: MoveKit Quality Hardening

**Defined:** 2026-06-21
**Core Value:** 安全、一键完成 C盘用户文件夹瘦身，迁移后所有软件正常工作，可随时回滚。

## v1 Requirements

Requirements for this quality-focused milestone.

### Bug Fixes

- [ ] **BUG-01**: `process.go:43` unsafe type assertion 修复 — 切换到 comma-ok 模式，防止 panic
- [ ] **BUG-02**: Junction debug buffer 大小修复 — 与 `REPARSE_DATA_BUFFER` 结构对齐
- [ ] **BUG-03**: Backup save 错误不再被静默吞掉 — 返回 MoveKitError 或回滚状态标记
- [ ] **BUG-04**: 同盘迁移降级方案 — 检测到同盘时直接 `os.Rename` 而非 copy+delete
- [ ] **BUG-05**: `dirSize` 错误传播 — 错误时返回 sentinel 值或 error，前端可感知
- [ ] **BUG-06**: Shell notify UTF16 编码错误处理 — 错误时日志记录，不阻塞主流程
- [ ] **BUG-07**: `move.go:137` percent 死代码清理 — 删除或接入实际进度计算
- [ ] **BUG-08**: 安全软件检测结果接入 — 检测结果在 PreCheck 返回中呈现

### CI/CD Foundation

- [ ] **CI-01**: GitHub Actions workflow on `windows-latest` — lint + test + build
- [ ] **CI-02**: `golangci-lint` 集成 — 覆盖所有 Go 包，阻塞 PR
- [ ] **CI-03**: `go test -race -cover` 自动运行，覆盖率报告输出
- [ ] **CI-04**: `wails build` 验证 — 确保每次提交可构建

### Go Backend Test Coverage

- [ ] **GO-01**: Syscall 接口抽象层 — RegistryAccessor / ProcessChecker / VolumeChecker 接口
- [ ] **GO-02**: `internal/syscall/` 全部 7 个源文件有对应测试（当前 6/7）
- [ ] **GO-03**: `internal/migrate/` 全部 17 个源文件有对应测试（当前 5/17）
- [ ] **GO-04**: `app.go` 17 个 Wails 绑定方法有测试覆盖（当前 0/17）
- [ ] **GO-05**: 测试覆盖率门禁 — 整体 ≥80%，migrate 包 ≥85%

### Frontend Quality

- [ ] **FE-01**: TypeScript strict 模式零错误
- [ ] **FE-02**: Vitest + React Testing Library 环境搭建
- [ ] **FE-03**: 12 个 React 组件全部有基础渲染测试
- [ ] **FE-04**: `App.tsx` 状态机逻辑有测试覆盖
- [ ] **FE-05**: `lib/wails.ts` mock 工厂可注入任意测试数据

### Integration & Verification

- [ ] **INT-01**: `Migrate()` 全流程集成测试（t.TempDir + StubRegistry）
- [ ] **INT-02**: `Rollback()` 集成测试（纯内存 fake，不动真实注册表）
- [ ] **INT-03**: Wails 绑定层冒烟测试（真实 Wails runtime）
- [ ] **INT-04**: 端到端迁移+回滚冒烟测试

### Improvements

- [ ] **IMP-01**: 错误处理审计 — 全局检查被忽略的 error 返回值
- [ ] **IMP-02**: 前端 `simulateProgress` 替换为真实 Wails 事件监听
- [ ] **IMP-03**: 取消按钮接入后端 `CancelMigration()` 调用
- [ ] **IMP-04**: `ResolveConflict` 桩实现 — 至少返回错误而非 nil

## v2 Requirements

Deferred to future release.

- **PERF-01**: 并行文件拷贝（多 goroutine 提速跨盘迁移）
- **PERF-02**: `dirSize` 缓存（避免重复 tree walk）
- **SEC-01**: `activeCopyCallback` 并发安全化
- **UX-01**: 真实 Windows Toast 通知（go-toast）
- **UX-02**: BitLocker 真实检测（非桩）
- **UX-03**: SSD 真实检测（IOCTL_STORAGE_QUERY_PROPERTY）
- **UX-04**: 更新检查接入 GitHub Releases API

## Out of Scope

| Feature | Reason |
|---------|--------|
| 新文件夹类型扩展 | 保持 5 个 Known Folder 不变 |
| UI 大改版 | 本轮聚焦质量，视觉翻新另起 phase |
| 安装包/代码签名 | 保持便携 exe 分发 |
| macOS/Linux 支持 | Windows only |
| E2E 测试（Playwright） | 研究建议此阶段跳过，手动冒烟足够 |
| 国际化/多语言 | 保持中文 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| BUG-01 ～ BUG-08 | Phase 0 | Pending |
| CI-01 ～ CI-04 | Phase 1 | Pending |
| GO-01 ～ GO-05 | Phase 2 | Pending |
| FE-01 ～ FE-05 | Phase 3 | Pending |
| INT-01 ～ INT-04 | Phase 4 | Pending |
| IMP-01 ～ IMP-04 | Phase 0 | Pending |

**Coverage:**
- v1 requirements: 30 total
- Mapped to phases: 30
- Unmapped: 0 ✓

---
*Requirements defined: 2026-06-21*
*Last updated: 2026-06-21 after initial definition*
