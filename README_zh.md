# PureCore

**基于 Go 语言的全栈 Web 开发框架 —— 类似 Laravel 开发风格，由 GoFiber v3 驱动。**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev/)

> ⚠️ 此项目目前处于 **Alpha** 阶段。API 和功能可能随时变动，恕不另行通知。

## 🚀 概述

PureCore 将 GoFiber v3 封装成类似 Laravel 的开发风格，提供路由分组、中间件管道、请求验证、统一响应格式、多语言支持和数据库 ORM 等开箱即用的功能。前端采用 Vue 3 + Vite + Tailwind CSS + DaisyUI 构建，支持服务端渲染。

- **后端**：Go · GoFiber v3 · PostgreSQL · GORM · go-playground/validator
- **前端**：Vue 3 · Vite · Tailwind CSS · DaisyUI · Bun（SSR）

## 📦 快速开始

```bash
git clone https://github.com/zhuchunshu/PureCore.git
cd PureCore
cp .env.example .env          # 配置数据库和端口
go build -o purecore .
./purecore migrate             # 运行数据库迁移
./purecore serve               # → http://localhost:9002

# 前端（新终端窗口）
cd web && bun install && bun run dev   # → http://localhost:9001
```

## 📚 文档

| 主题 | 中文 | EN |
|------|------|-----|
| 框架指南 | [docs/zh/README.md](docs/zh/README.md) | [docs/en/README.md](docs/en/README.md) |
| CLI 命令 | [docs/zh/CLI.md](docs/zh/CLI.md) | [docs/en/CLI.md](docs/en/CLI.md) |
| 数据库与模型 | [docs/zh/DATABASE.md](docs/zh/DATABASE.md) | [docs/en/DATABASE.md](docs/en/DATABASE.md) |
| API 参考 | [docs/zh/API.md](docs/zh/API.md) | [docs/en/API.md](docs/en/API.md) |
| SSR 指南 | [docs/zh/SSR.md](docs/zh/SSR.md) | [docs/en/SSR.md](docs/en/SSR.md) |
| 开发指南 | [docs/zh/DEVELOPMENT.md](docs/zh/DEVELOPMENT.md) | [docs/en/DEVELOPMENT.md](docs/en/DEVELOPMENT.md) |

## 📄 许可证

PureCore 基于 [MIT 许可证](LICENSE) 开源。

---

**作者**：[zhuchunshu](https://github.com/zhuchunshu) · **仓库**：[GitHub](https://github.com/zhuchunshu/PureCore)
