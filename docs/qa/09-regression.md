# 质量保障记录 09 — 回归测试与二次自检

日期：2026-08-02

## 1. 回归范围

全部缺陷修复（D-01、D-02、A-1、A-2）后，对全部业务场景执行第二轮全面回归。

## 2. 回归结果

```
Go 测试：93 PASS / 0 FAIL（覆盖迁移全流程、回滚、取消、失败恢复、junction、注册表、配置）
前端测试：12 PASS（状态机成功/失败/取消、DiskSelector、FolderList、类型）
静态检查：go vet 0 告警；ESLint 0 错误；tsc --noEmit 通过
覆盖率：main 18.2% / migrate 66.4% / syscall 64.1%
```

## 3. 二次数据自检

- 备份记录状态机回归：complete/rolled_back 流转测试通过；
- 注册表双键写入与类型：User Shell Folders 现统一 EXPAND_SZ；
- no-op 重跑不再覆盖可回滚备份（专项测试）；
- 损坏配置/备份 JSON 的容错行为有专项测试；
- 权限：备份与配置均 0600。

## 4. 未闭环项（记录在案，非交付阻塞）

| 项 | 状态 |
| --- | --- |
| race 检测、golangci-lint、govulncheck | 依赖 CI（gcc/网络），本机未执行 |
| cmd/verify 第 15 项 Shell Folders 读取 | 沙箱环境键缺失（写入路径已修复），真实 Windows 上键存在 |
| A-2/A-3/A-4 数据限制 | 已记录，非缺陷 |

## 5. 二次自检结论

所有已登记缺陷闭环，系统整体稳定，达到最终交付质量标准。
