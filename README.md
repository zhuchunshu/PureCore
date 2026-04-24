# PureCore

**A full-stack Go web development framework — Laravel-like style, powered by GoFiber v3.**

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev/)

> ⚠️ This project is currently in **Alpha** stage. APIs and features may change without notice.

## 🚀 Overview

PureCore wraps GoFiber v3 into a Laravel-like development style, providing routing groups, middleware pipelines, request validation, unified response formatting, and multi-language support out of the box. The frontend is built with Vue 3 + Vite + Tailwind CSS + DaisyUI.

- **Backend**: Go · GoFiber v3 · PostgreSQL · GORM · go-playground/validator
- **Frontend**: Vue 3 · Vite · Tailwind CSS · DaisyUI · Bun

## ✨ Features

| Feature | Description |
|---------|-------------|
| **Declarative Routing** | Chainable group, prefix, and middleware definitions — code as documentation |
| **Intelligent Validation** | Built-in validation engine with common rules, one line to secure inputs |
| **Native i18n** | Frontend and backend share one set of language resources with auto-detection |
| **Standardized Output** | Unified JSON response structure with built-in pagination and error codes |

## 📦 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL
- Bun (or Node.js + npm)
- Git

### Setup

```bash
# Clone
git clone https://github.com/zhuchunshu/PureCore.git
cd PureCore

# Configure environment
cp .env.example .env
# Edit .env with your database credentials and ports

# Start backend
go mod tidy
ln -s ../../lang web/public/lang
go run main.go         # → http://localhost:9002

# Start frontend
cd web
bun install
bun run dev            # → http://localhost:9001
```

## 📁 Project Structure

```
/purecore
├── core/                  # Core framework (router, request, response, lang)
├── app/Http/
│   ├── Controllers/       # Application controllers
│   └── Middleware/         # Auth, CORS, Language detection
├── routes/                # Route registration
├── lang/                  # Shared translation files (zh.json, en.json)
├── web/                   # Vue 3 frontend
├── docs/                  # Documentation (EN & ZH)
├── LICENSE                # MIT License
├── .env.example           # Environment template
├── main.go                # Entry point
└── go.mod
```

## 🌐 API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/v1/ping` | No | Health check |
| GET | `/api/v1/system/info` | No | Project information |
| GET | `/api/v1/users` | Yes | List users |
| POST | `/api/v1/users` | Yes | Create user |
| GET | `/api/v1/users/:id` | Yes | Get user |

Authentication: `Authorization: Bearer <token>`

## 📚 Documentation

- [English Docs](docs/en/README.md)
- [中文文档](docs/zh/README.md)
- [API Reference (EN)](docs/en/API.md)
- [Development Guide (EN)](docs/en/DEVELOPMENT.md)

## ⚙️ Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `FRONTEND_PORT` | Frontend dev server port | `9001` |
| `BACKEND_PORT` | Backend API server port | `9002` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `purecore` |
| `APP_ENV` | Runtime environment | `local` |
| `APP_DEBUG` | Debug mode | `true` |

## 🤝 Contributing

Contributions are welcome! Please read the [Development Guide](docs/en/DEVELOPMENT.md) to get started.

## 📄 License

PureCore is open-sourced under the [MIT license](LICENSE).

---

**Author**: [zhuchunshu](https://github.com/zhuchunshu) · **Repository**: [GitHub](https://github.com/zhuchunshu/PureCore)
