# PureCore

**基于 Go 语言的全栈 Web 开发框架 —— 类似 Laravel 开发风格，由 GoFiber v3 驱动。**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev/)
[![Version](https://img.shields.io/badge/version-1.0.0_alpha-orange.svg)](purecore.json)

> ⚠️ 此项目目前处于 **Alpha** 阶段。API 和功能可能随时变动，恕不另行通知。

## 🚀 概述

PureCore 将 GoFiber v3 封装成类似 Laravel 的开发风格，提供路由分组、中间件管道、请求验证、统一响应格式和多语言支持等开箱即用的功能。前端采用 Vue 3 + Vite + Tailwind CSS + DaisyUI 构建。

- **后端**：Go · GoFiber v3 · PostgreSQL · GORM · go-playground/validator
- **前端**：Vue 3 · Vite · Tailwind CSS · DaisyUI · Bun

## ✨ 核心特性

| 特性 | 描述 |
|------|------|
| **声明式路由** | 链式定义分组、前缀和中间件 —— 代码即文档 |
| **智能验证** | 内置验证引擎，覆盖常用规则，一行代码完成输入校验 |
| **原生国际化** | 前后端共享一套语言资源，自动检测语言 |
| **标准化输出** | 统一的 JSON 响应结构，内建分页和错误码 |

## 📦 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL
- Bun（或 Node.js + npm）
- Git

### 安装步骤

```bash
# 克隆项目
git clone https://github.com/zhuchunshu/PureCore.git
cd PureCore

# 配置环境
cp .env.example .env
# 编辑 .env 文件，设置数据库连接和端口

# 启动后端
go mod tidy
ln -s ../../lang web/public/lang
go run main.go         # → http://localhost:9002

# 启动前端
cd web
bun install
bun run dev            # → http://localhost:9001
```

## 📁 项目结构

```
/purecore
├── core/                  # 核心框架（路由、请求、响应、语言）
├── app/Http/
│   ├── Controllers/       # 应用控制器
│   └── Middleware/         # 鉴权、跨域、语言检测
├── routes/                # 路由注册
├── lang/                  # 共享翻译文件（zh.json、en.json）
├── web/                   # Vue 3 前端
├── docs/                  # 文档（英文和中文）
├── purecore.json          # 项目元数据
├── .env.example           # 环境配置模板
├── main.go                # 入口文件
└── go.mod
```

## 🌐 API 接口

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| GET | `/api/v1/ping` | 否 | 健康检查 |
| GET | `/api/v1/system/info` | 否 | 项目信息 |
| GET | `/api/v1/users` | 是 | 用户列表 |
| POST | `/api/v1/users` | 是 | 创建用户 |
| GET | `/api/v1/users/:id` | 是 | 用户详情 |

认证方式：`Authorization: Bearer <token>`

## 📚 文档

- [中文文档](docs/zh/README.md)
- [English Docs](docs/en/README.md)
- [API 参考 (中文)](docs/zh/API.md)
- [开发指南 (中文)](docs/zh/DEVELOPMENT.md)

## ⚙️ 配置说明

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `FRONTEND_PORT` | 前端开发服务器端口 | `9001` |
| `BACKEND_PORT` | 后端 API 服务器端口 | `9002` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `DB_USER` | 数据库用户名 | `postgres` |
| `DB_PASSWORD` | 数据库密码 | `postgres` |
| `DB_NAME` | 数据库名称 | `purecore` |
| `APP_ENV` | 运行环境 | `local` |
| `APP_DEBUG` | 调试模式 | `true` |

## 🤝 贡献

欢迎贡献！请阅读[开发指南](docs/zh/DEVELOPMENT.md)开始。

## 📄 许可证

PureCore 基于 [MIT 许可证](LICENSE) 开源。

---

**作者**：[zhuchunshu](https://github.com/zhuchunshu) · **仓库**：[GitHub](https://github.com/zhuchunshu/PureCore)
