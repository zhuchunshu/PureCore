## Brief overview
This file contains guidelines for developing on the **PureCore** project — a full-stack Go web framework using GoFiber v3 + PostgreSQL (backend) and Bun + Vite + Vue 3 + TailwindCSS + DaisyUI (frontend).

## Project architecture
- Backend entry: `main.go` → Cobra CLI (`cmd/`) → GoFiber v3 server
- Core layer (`core/`) provides: database singleton, request/response wrappers, router builder, JWT middleware helpers, language manager, migration runner
- Routes defined in `routes/` directory using Laravel-style chaining: `Prefix → Middleware → Group`
- **IMPORTANT**: Backend routes MUST be defined in files under `routes/`, NOT in `RouteServiceProvider.go`. Create a new file (e.g., `routes/sessions.go`) for each feature group and call its registration function from `cmd/serve.go` after `app.Boot()`.
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

## Dashboard page splitting rules
- Each dashboard feature (profile, security, api keys, sessions, etc.) MUST be a **separate page component** under `web/src/pages/dashboard/`, e.g. `ProfilePage.vue`, `SecurityPage.vue`, `SessionsPage.vue`
- `UserDashboard.vue` MUST act as a **layout container** only — it provides the sidebar navigation and a `<router-view />` for child pages. It MUST NOT directly include the feature components.
- Each dashboard child page MUST have a **unique URL route** (e.g. `/dashboard/profile`, `/dashboard/security`, `/dashboard/sessions`)
- Sidebar items MUST use `<router-link>` to navigate between child routes, enabling direct URL sharing and browser back/forward support
- Route definitions MUST use nested `children` arrays under the `/dashboard` parent route

## Frontend conventions
- Vue 3 with `<script setup>` syntax and Composition API
- Styling: TailwindCSS 4 utility classes + DaisyUI 5 component classes
- Routing: Vue Router 5 with SSR (memory history on server, web history on client)
- State management: Vue composables (`web/src/composables/`)
- API calls: proxy `/api` to backend (port 9002) via Vite dev server or SSR server
- Environment variables: `VITE_ADMIN_ROUTE_PREFIX` for admin route prefix

## Page background color standard
- **ALL non-homepage pages MUST use `bg-base-200` as the background color** for their main content area (excluding Navbar and Footer). This ensures a consistent, eye-friendly off-white/dark-gray background across all pages.
- The DocsPage (`web/src/pages/DocsPage.vue`) uses `bg-base-200` — all other pages must match this exactly.
- **Exceptions (NEVER modify these pages' backgrounds):**
  - **HomePage** — exempt because it has its own hero section with CyberBackground.
  - **NotFound (404 page)** — MUST keep `bg-base-100` (pure white / theme default). NEVER change this.
- Examples of correct usage:
  - `<div class="min-h-screen bg-base-200">` for standalone pages (LoginPage, RegisterPage, etc.)
  - `<div class="flex-1 bg-base-200">` for pages wrapped in a layout (UserDashboard, Admin pages via AdminLayout, etc.)
  - AdminLayout already uses `bg-base-200` as its wrapper background
- Navbar and Footer should retain their own backgrounds (typically `bg-base-100`) and should NOT be affected by this rule.

## SSR safety rules (CRITICAL)
- **NEVER** access `window`, `document`, `localStorage`, `navigator`, or any other browser-only API directly in Vue templates (`<template>`) or in `<script setup>` top-level scope. These do not exist during server-side rendering and will cause the component to crash, preventing hydration — the page will stay in an infinite loading state.
- Always guard browser API access with `typeof window !== 'undefined'` check inside `onMounted()`, and store the result in a `ref()` for use in the template.
- Example (correct):
  ```js
  const callbackUrl = ref('')
  onMounted(() => {
    if (typeof window !== 'undefined') {
      callbackUrl.value = `${window.location.protocol}//${window.location.host}/callback`
    }
  })
  ```
  Then use `:value="callbackUrl"` in the template instead of `` :value="`${window.location.protocol}//...`" ``.
- **NEVER** do this in a template: `:value="window.location.href"` or `{{ document.title }}` — these will crash SSR.

## Loading spinner rules (CRITICAL)
- When a page component uses a loading spinner pattern (`loading` ref starts as `true`, set to `false` in `onMounted`), every early return path in `onMounted` MUST set `loading.value = false` before returning. Failure to do so will leave the spinner running forever.
- Example (correct):
  ```js
  onMounted(async () => {
    if (!accessToken.value) {
      loading.value = false  // REQUIRED before early return
      router.push('/login')
      return
    }
    try { ... } finally { loading.value = false }
  })
  ```

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

## API call guard rule
- If an API endpoint requires authentication (login) to be useful, the frontend **must** avoid calling it when the user is not logged in. Always check the authentication state before making such requests.

## Icon library
- **NEVER** use emoji for icons anywhere in the frontend
- Primary icon library: `lucide-vue-next` (Lucide icons)
- Install: `bun add lucide-vue-next`

- Usage example (Lucide):
  ```vue
  <script setup>
  import { Settings, User, LogOut } from 'lucide-vue-next'
  </script>
  <template>
    <Settings :size="20" />
  </template>
  ```

## Language / i18n
- Backend loads translations from `lang/` directory (JSON files, nested structure)
- Frontend i18n module loads translations via `/lang/{locale}.json` (proxied from `web/public/lang/`)
- Cookie `purecore-locale` persists user language choice across requests
- Both sides flatten nested JSON to `"group.key"` format
