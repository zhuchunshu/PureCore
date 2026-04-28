# PureCore

**A full-stack Go web application framework — Laravel-like style, powered by GoFiber v3.**

[[License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[[Go Version](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://go.dev/)
[[Template](https://img.shields.io/badge/github-template-blue.svg)](https://github.com/zhuchunshu/PureCore)

> ⚠️ This project is currently in **Alpha** stage. APIs and features may change without notice.

## 🚀 Overview

PureCore wraps GoFiber v3 into a Laravel-like development style, providing routing groups, middleware pipelines, request validation, unified response formatting, multi-language support, and database ORM out of the box. The frontend is built with Vue 3 + Vite + Tailwind CSS + DaisyUI with SSR support.

- **Backend**: Go · GoFiber v3 · PostgreSQL · GORM · go-playground/validator
- **Frontend**: Vue 3 · Vite · Tailwind CSS · DaisyUI · Bun (SSR)

## 📦 Use This Template

This is a **GitHub template repository** — click the green "Use this template" button at the top of the repo to create your own project from it.

### 1️⃣ Create Your Project

```bash
# Option A: Via GitHub's "Use this template" button (recommended)
# Clone your new repo, then run the setup wizard:
chmod +x scripts/setup.sh
./scripts/setup.sh
```

The wizard will ask for:
- Your project name (Go module name)
- Custom ports (frontend, backend, database)
- Database credentials (generates secure random passwords)
- Admin route prefix and default theme
- Then automatically renames the Go module, updates all imports,
  initializes a fresh Git repo, and builds the project

### 2️⃣ Start Developing

```bash
# Backend
go build -o purecore .
./purecore serve               # → http://localhost:9002

# Frontend (new terminal)
cd web && bun install && bun run dev   # → http://localhost:9001
```

### 3️⃣ Deploy to Production

```bash
# Interactive deployment (asks for ports)
./purecore deploy
# Or directly:
chmod +x scripts/deploy.sh && ./scripts/deploy.sh
```

## 📚 Documentation

| Topic | EN | 中文 |
|-------|----|------|
| Framework Guide | [docs/en/README.md](docs/en/README.md) | [docs/zh/README.md](docs/zh/README.md) |
| CLI Commands | [docs/en/CLI.md](docs/en/CLI.md) | [docs/zh/CLI.md](docs/zh/CLI.md) |
| Database & Models | [docs/en/DATABASE.md](docs/en/DATABASE.md) | [docs/zh/DATABASE.md](docs/zh/DATABASE.md) |
| API Reference | [docs/en/API.md](docs/en/API.md) | [docs/zh/API.md](docs/zh/API.md) |
| SSR Guide | [docs/en/SSR.md](docs/en/SSR.md) | [docs/zh/SSR.md](docs/zh/SSR.md) |
| Development Guide | [docs/en/DEVELOPMENT.md](docs/en/DEVELOPMENT.md) | [docs/zh/DEVELOPMENT.md](docs/zh/DEVELOPMENT.md) |

## 📄 License

PureCore is open-sourced under the [MIT license](LICENSE).

---

**Author**: [zhuchunshu](https://github.com/zhuchunshu) · **Repository**: [GitHub](https://github.com/zhuchunshu/PureCore)
