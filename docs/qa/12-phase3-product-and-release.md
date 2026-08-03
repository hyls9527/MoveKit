# 第三阶段产品化与 v0.3.0 发布记录

日期：2026-08-02 ~ 2026-08-03
范围：产品化功能 + 审查修复 + 发布（PR #21、#22、#23）

## 产品化交付

| 功能 | 说明 |
| --- | --- |
| 自动更新 | CheckUpdate 查询 GitHub Releases API，前端"检查更新"按钮 + 下载横幅 |
| 迁移成果反馈 | 完成页展示本次释放空间 + C 盘当前可用空间 |
| 代码签名框架 | CI 集成 Azure Trusted Signing（凭据未配置时自动跳过），SIGNING.md 文档 |

## 审查修复（PR #22）

- Rollback 白名单校验（阻止路径穿越）
- 版本号单源（去掉 ldflags 注入 + release 版本一致性校验）
- 迁移生命周期（shutdown 取消并等待回滚）
- main.go 失败退出码 0 -> 1
- 删除 main 包 MoveKitError/MigrationStatus/DiskDetail 死类型
- 本机 Go 升级 1.26.4 -> 1.26.5，漏洞扫描归零

## 发布

- v0.3.0 tag 指向 `3e044a6`，CI lint/vuln/test/build/release 全绿
- Release v0.3.0 已发布（MoveKit.exe 13.4 MB），latest 指向 v0.3.0
- 版本一致性校验首次实战生效（错误 tag 被拦截后重建）

## 结论

项目达到收尾质量标准，无遗留阻塞项。
