# MoveKit — C盘瘦身工具

[![CI](https://github.com/hyls9527/MoveKit/actions/workflows/ci.yml/badge.svg)](https://github.com/hyls9527/MoveKit/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/hyls9527/MoveKit)](https://github.com/hyls9527/MoveKit/releases/latest)

**一键迁移桌面/下载/文档/图片/音乐到其他盘，释放 C 盘空间。所有程序无需任何改动，照常运行。**

---

## 为什么需要 MoveKit

C 盘空间不足时，Windows 变慢、更新失败、软件报错。手动迁移用户文件夹繁琐且危险——改错注册表或路径会导致程序找不到文件。

MoveKit 把这件事变成一次点击：选盘→确认→完成。迁移后回到 C 盘，一切都和原来一样。

## 快速开始

1. **下载** — 从 [Releases](https://github.com/hyls9527/MoveKit/releases) 获取最新 `MoveKit.exe`
2. **运行** — 双击启动（无需安装）
3. **选盘** — 选择一个 NTFS 磁盘作为目标
4. **迁移** — 点击「开始迁移」，等待完成

> 首次运行可能触发 SmartScreen，这是未签名便携 exe 的正常提示。点击「更多信息」→「仍要运行」即可。

## 功能亮点

| | |
|---|---|
| **一键搬家** | 桌面、下载、文档、图片、音乐，5 个文件夹同时迁移 |
| **程序零感知** | 自动创建 NTFS Junction，所有快捷方式和软件路径不受影响 |
| **随时回滚** | 迁移前自动备份，一键恢复到原始位置 |
| **智能安检** | 检测 OneDrive 冲突、磁盘空间、文件占用、杀软拦截，迁移前预警 |

## 安全机制

- **备份优先** — 迁移前写入备份记录，所有操作可追溯
- **原子回滚** — 迁移失败时自动把文件移回原处，不留孤儿文件
- **双键注册表** — User Shell Folders + Shell Folders 同时更新，任一失败自动回滚
- **冲突拦截** — 迁移前检测目标盘同名文件，存在冲突时阻止迁移，避免覆盖数据
- **进程防重** — 单实例互斥锁，防止同时运行多个迁移

## 适用场景

- C 盘空间不足，用户文件夹占用过大
- 新电脑或重装系统后，希望把数据放到其他盘
- 从 HDD 迁移到 SSD，提升系统响应速度

## 技术栈

`Go 1.23.5` · `Wails v2` · `React 19` · `TypeScript` · `Tailwind CSS v4` · `Win32 API` · `Windows 10/11`

## 开发

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

## 协议

MIT License
