## Brief overview
This file contains guidelines for developing on the **PureCore** project — a full-stack Go web framework using GoFiber v3 + PostgreSQL (backend) and Bun + Vite + Vue 3 + TailwindCSS + DaisyUI (frontend).

## Project architecture
- Backend entry: `main.go` → Cobra CLI (`cmd/`) → GoFiber v3 server
- Core layer (`core/`) provides: database singleton, request/response wrappers, router builder, JWT middleware helpers, language manager, migration runner
- Routes defined in `routes/api.go` using Laravel-style chaining: `Prefix → Middleware → Group`
- Models in `app/Models/` embed `core.Model` (ID, CreatedAt, UpdatedAt, DeletedAt)
- Controllers in `app/Http/Controllers/` use `core.H()` wrapper that auto-injects `*core.Request` and `*core.Response`
- Frontend entry: `web/` — Vite SSR with `server.js` as the Node/Bun SSR server
- Shared translations in `lang/` (used by both Go backend and Vue frontend)

## Database and migrations
- All database schema changes MUST be done via migration files only
- Use `purecore make:migration <name>` to generate a new migration file in `database/migrations/`
- Never modify existing migration files — always create new ones
- Migrations are registered via `init()` and tracked in a `migrations` table in the database
- Models should extend `core.Model` for consistent base fields

## Frontend layout rules
- All admin pages after login MUST use `AdminLayout`/`AdminNavbar` — never use the public-facing `Navbar.vue`
- Admin pages include: AdminDashboard, AdminSettings, and any future pages under the `/{admin}/` route prefix
- Public pages (HomePage, NotFound, etc.) use `Navbar.vue` + `Footer.vue`
- Login and Register pages use the public `Navbar.vue` + `Footer.vue` — they are NOT standalone pages

## Frontend conventions
- Vue 3 with `<script setup>` syntax and Composition API
- Styling: TailwindCSS 4 utility classes + DaisyUI 5 component classes
- Routing: Vue Router 5 with SSR (memory history on server, web history on client)
- State management: Vue composables (`web/src/composables/`)
- API calls: proxy `/api` to backend (port 9002) via Vite dev server or SSR server
- Environment variables: `VITE_ADMIN_ROUTE_PREFIX` for admin route prefix

## CLI commands (code generation)
- `purecore serve` — start the HTTP server
- `purecore migrate` — run pending migrations
- `purecore make:model <Name>` — create a GORM model
- `purecore make:controller <Name>` — scaffold a CRUD controller
- `purecore make:migration <name>` — create a timestamped migration file

## Migration and model/controller synchronization
- When creating a new migration file, you MUST also create or update the corresponding model and controller
- Migration → Model → Controller should always stay in sync
- For example: creating a `create_web_options_table` migration means also creating `WebOption` model and `OptionController` with CRUD endpoints
- New models embed `core.Model` for consistent ID, timestamps, and soft delete fields

## Response format
All API endpoints return JSON in this shape:
```json
{ "code": 0, "message": "success", "data": ... }
```
Error responses: `{ "code": 4xx/5xx, "message": "error description" }`
Paginated responses include: `total`, `page`, `per_page` alongside `data`

## SEO (Search Engine Optimization)
- `web/src/composables/useSEO.js` provides a Vue composable for setting page-level SEO meta tags
- `useSEO({ title, description, keywords })` sets `<title>`, `<meta name="description">`, and `<meta name="keywords">` tags
- Title format: `"Page Title - Site Name"` or just `"Site Name"` if page title is empty
- Site name, description, and keywords are loaded from admin options (`site_name`, `site_description`, `site_keywords`) via `initSEO()`
- `initSEO()` is called in `web/src/main.js` at app startup alongside `initI18n()`
- SSR meta tags are injected in `web/server.js` using project info from `web/package.json` — placeholders `<!--seo-title-->`, `<!--seo-description-->`, `<!--seo-keywords-->` in `web/index.html` are replaced at render time
- Every page component should call `useSEO()` with page-specific title and description

## Authentication
- Admin routes use JWT with access token (15 min) + refresh token (7 days)
- Admin route prefix is configurable via `ADMIN_ROUTE_PREFIX` env var
- `AdminAuth` middleware verifies the token and checks `token_version` for invalidation
- `useAuth.js` composable handles client-side token management and auto-refresh

## Language / i18n
- Backend loads translations from `lang/` directory (JSON files, nested structure)
- Frontend i18n module loads translations via `/lang/{locale}.json` (proxied from `web/public/lang/`)
- Cookie `purecore-locale` persists user language choice across requests
- Both sides flatten nested JSON to `"group.key"` format
