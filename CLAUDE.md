# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Backend
go build -o purecore .
./purecore serve                    # Start server (default :9002)
./purecore migrate                  # Run database migrations
./purecore deploy                   # Interactive deployment

# Frontend (SSR dev mode)
cd web && bun install && bun run dev    # → :9001

# Frontend (SPA-only dev)
cd web && bun run dev:spa

# Production build
cd web && bun run build             # Client + SSR server bundle
cd web && bun run preview           # Preview production build
```

## Architecture Overview

**Stack**: GoFiber v3 (backend) + Vue 3 / Vite / Tailwind CSS / DaisyUI (frontend SSR with Bun)

**Module name**: `purecore` (Go 1.26.2)

### Backend Core (`core/`)

The framework kernel. Key files:

- **`application.go`** — `Application` container: bootstraps config, database, language, routing, and service providers. The main entry point for the HTTP server.
- **`router.go`** — Fluent Laravel-style router: `.Prefix()`, `.Middleware()`, `.Group(fn)`, `.RegisterNamedMiddleware()`. Wraps Fiber's router.
- **`middleware.go`** — Defines `HandlerFunc func(req *Request, res *Response) error` and the `H()` adapter that converts it to `fiber.Handler`. All controllers use this signature. Also aliases `MiddlewareFunc = fiber.Handler`.
- **`service_provider.go`** — `ServiceProvider` interface (`Name/Register/Boot`). All route registration goes through providers.
- **`request.go`** — `Request` wraps `fiber.Ctx` with `Input(key)`, `Validate(struct)`, `User()`, `BearerToken()`.
- **`response.go`** — `Response` wraps `fiber.Ctx` with `Success()`, `Error()`, `Paginate()`, `ValidationError()`, `Created()`, `NoContent()`. All responses use the `{code, message, data}` JSON format.
- **`database.go`** — Singleton GORM PostgreSQL connection via `DB()`.
- **`config.go`** — Singleton `Config` reading from env vars (`APP_DEBUG`, `DB_HOST`, etc.).
- **`lang.go`** — i18n via JSON files in `lang/`. Nested keys flattened with dot notation (`common.success`).
- **`option.go`** — Key-value store backed by `web_options` DB table with in-memory cache.
- **`migrator.go`** — Compile-time migration system: files in `database/migrations/` register via `init()` → `core.RegisterMigration()`. Runs on boot.
- **`model.go`** — Base `Model` struct (ID, CreatedAt, UpdatedAt, soft delete).
- **`errors.go`** — `AppError` and `ValidationError` types with common HTTP error sentinels.
- **`turnstile.go`** — Cloudflare Turnstile verification helpers.

### Request Flow

1. `cmd/serve.go` creates `core.NewApplication()`, adds providers, calls `Boot()`.
2. `Boot()` initializes lang, DB, runs migrations, creates Fiber app + router, registers providers.
3. `RouteServiceProvider` registers all API routes with named middleware (`auth`, `admin_auth`, `cors`, `lang`).
4. Global middleware (`Cors`, `Lang`) applied before `app.Run()`.
5. Route handlers use `core.H()` adapter: `func(req *core.Request, res *core.Response) error`.

### Route Structure

- `GET/POST /api/v1/*` — Public API (ping, system info, auth register/login/refresh, docs)
- `GET/POST /api/v1/{admin_prefix}/*` — Admin routes (auth check/login/register/refresh, options). Authenticated admin routes under middleware `AdminAuth()`.
- `GET/POST/PUT/DELETE /api/v1/*` with `Auth()` middleware — Authenticated user routes (profile, user CRUD)
- Session routes registered separately in `routes/sessions.go`

### Frontend (`web/`)

- **SSR architecture**: `entry-server.js` → `entry-client.js`, served by `server.js` (Bun/Vite dev server).
- **Pages**: `LoginPage`, `RegisterPage`, `UserDashboard`, `HomePage`, `DocsPage`, `NotFound`, plus admin pages in `pages/admin/`.
- **Composables**: `useAuth` (admin), `useUserAuth` (user), `useAdminOption`, `useTheme`, `useToast`, `useAvatar`, `useSEO`, `useBackendHealth`, `useUserAuth`.
- **State**: No Pinia/Vuex — composables use module-level reactive state.
- **Styling**: Tailwind CSS 4 + DaisyUI 5.
- **API calls**: `services/api.js` (fetch wrapper with auth token injection).
- **i18n**: `i18n.js` uses `navigator.language`.

### Key Patterns

- **Adding a new API route group**: Create a new file in `routes/` (e.g., `routes/sessions.go`), call its registration function from `cmd/serve.go` after `app.Boot()`. Do NOT add routes directly in `RouteServiceProvider.Register()`.
- **Adding a new API endpoint (simple)**: Add route in `RouteServiceProvider.Register()`, create controller in `app/Http/Controllers/`.
- **Adding a new page**: Create `.vue` in `web/src/pages/`, add route in `web/src/router/routes.js`.
- **Database migrations**: Use `./purecore make:migration <name>`. Never modify existing migration files — always create new ones. Call `core.RegisterMigration()` in `init()`, blank-import in `cmd/serve.go`. When creating a migration, also create/update the corresponding Model and Controller.
- **Configuration**: All config via `.env` file, accessed through `core.GetConfig()`.
- **Middleware**: Create in `app/Http/Middleware/`, register by name in `RouteServiceProvider.Register()`.

### API Response Format

All endpoints return `{code, message, data}` JSON. Paginated responses include `total`, `page`, `per_page`.

### Frontend Conventions (CRITICAL)

- **SSR Safety**: NEVER access `window`, `document`, `localStorage`, `navigator` in `<template>` or top-level `<script setup>`. Always guard with `typeof window !== 'undefined'` inside `onMounted()`, store result in `ref()`. SSR crashes cause infinite loading.
- **Loading States**: NEVER use `loading loading-spinner` for page-level loading. Always use **DaisyUI skeleton** components that mirror the page structure. Exception: button inline spinners (`loading-spinner-xs`) are allowed during form submission.
- **Loading Guard**: Every early return path in `onMounted` MUST set `loading.value = false` before returning, or the skeleton stays forever.
- **Page Background**: All non-homepage pages use `bg-base-200` for main content. Exceptions: HomePage (hero), NotFound (`bg-base-100` — never change).
- **Dashboard Pages**: Each feature is a separate page component under `web/src/pages/dashboard/` with unique URL routes. `UserDashboard.vue` is a layout container only — sidebar + `<router-view />`, nested `children` routes.
- **Admin Layout**: All admin pages after login MUST use `AdminLayout`/`AdminNavbar`. Public pages use `Navbar.vue` + `Footer.vue`.
- **Icons**: NEVER use emoji. Primary: `lucide-vue-next`. If Lucide doesn't have the desired icon, fall back to `@tabler/icons-vue`.
- **Dependencies**: Third-party libraries are allowed when they solve a real problem — no need to reinvent the wheel.
- **API Call Guard**: If an API endpoint requires authentication, check auth state before calling it.
- **SEO**: `useSEO({ title, description, keywords })` composable for meta tags. `initSEO()` called in `main.js`.

### Admin vs User Auth

Two separate auth systems:
- **Admin**: `AdminAuth` middleware + `AdminAuthController` (JWT-based, stored in `admin_users` table)
- **User**: `Auth` middleware + `UserAuthController` (JWT-based, stored in `users` table)
- Each has its own `/auth/login`, `/auth/register`, `/auth/refresh`, `/auth/profile` endpoints with separate route prefixes.
