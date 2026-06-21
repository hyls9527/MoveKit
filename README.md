# MoveKit - C盘瘦身工具

[![CI](https://github.com/hyls9527/MoveKit/actions/workflows/ci.yml/badge.svg)](https://github.com/hyls9527/MoveKit/actions/workflows/ci.yml)

一键迁移用户文件夹（桌面/下载/文档/图片/音乐）到其他盘，安全可靠。

## ✨ 功能

- **一键批量迁移** — 桌面、下载、文档、图片、音乐，5 个文件夹一键搬家
- **100% 兼容** — 自动创建 NTFS Junction，所有程序照常使用
- **安全可靠** — 迁移前自动备份，一键回滚到原始位置
- **实时进度** — 显示文件数、字节数、百分比三重进度
- **智能检测** — OneDrive 冲突、文件占用、磁盘空间、杀软拦截
- **便携版** — 单个 exe，双击即用，无需安装

## 📦 下载

从 [Releases](../../releases) 页面下载最新版 `MoveKit.exe`。

## 🚀 使用

1. 双击 `MoveKit.exe` 运行
2. 选择目标磁盘（推荐 SSD）
3. 点击"开始迁移"
4. 等待完成

## ⚠️ 注意事项

- 首次运行可能触发 Windows Defender / SmartScreen 警告，这是误报。点击"更多信息" → "仍要运行"
- 如果正在使用 OneDrive 同步，请先关闭文件夹同步再迁移
- 迁移过程中请勿关闭程序
- 迁移完成后可随时回滚

## 🔧 开发

### 环境要求

- Go 1.22+
- Node.js 18+
- Wails v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### 开发模式

```bash
cd MoveKit
wails dev
```

### 构建

```bash
cd MoveKit/frontend && npx vite build
cd .. && go build -o MoveKit.exe .
```

### 测试

```bash
go test ./... -v
```

## 📁 项目结构

```
MoveKit/
├── main.go                    # Wails 入口
├── app.go                     # 16 个 Wails 绑定方法
├── app_types.go               # 类型定义
├── internal/
│   ├── syscall/               # Win32 API 封装 (junction/registry/volume/fileops/process/power/shell)
│   ├── migrate/               # 迁移核心逻辑 (migrate/check/onedrive/backup/rollback/move/conflict/state/verify/security/cloud/version/errors/log)
│   └── folders/               # Known Folder GUID 定义
└── frontend/                  # React + TypeScript + TailwindCSS v4
    └── src/
        ├── App.tsx            # 状态机 + 所有对话框集成
        └── components/        # 12 个 UI 组件
```

## 📄 协议

MIT License
