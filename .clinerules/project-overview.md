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

## Loading state rules (CRITICAL) — DaisyUI Skeleton
- **NEVER use loading spinners** (`loading loading-spinner`) for full-page or component-level initial loading states. Always use **DaisyUI skeleton** components instead.
- Skeleton components provide a much better user experience by visually reflecting the page structure while data loads, rather than showing a blank screen with a spinning circle.
- Reference: https://daisyui.com/components/skeleton/
- **Exception**: Button-level inline spinners (e.g., `<span class="loading loading-spinner loading-xs"></span>` inside a `<button>` during form submission) are still acceptable since they indicate an action in progress, not a page load.

### How to design a skeleton
1. Look at the **final rendered layout** of the page/component when data is loaded.
2. Create a `<template v-if="loading">` block that mirrors the structural hierarchy using `<div class="skeleton ...">` elements.
3. Match skeleton dimensions (height, width, rounded corners) to approximate the real content.
4. Use `skeleton` class combined with Tailwind sizing utilities:
   - `skeleton h-28 rounded-2xl` — header area
   - `skeleton h-4 w-3/4 rounded` — text line
   - `skeleton h-10 w-32 rounded-xl` — tab/button
   - `skeleton w-10 h-10 rounded-lg` — icon placeholder
   - `skeleton h-96 rounded-2xl` — card/table area

### Skeleton examples by page type

**Dashboard overview page** (profile card + stat cards + resource cards):
```html
<template v-if="loading">
  <!-- Profile card skeleton -->
  <div class="skeleton h-24 rounded-2xl"></div>
  <!-- Stat cards skeleton -->
  <div class="grid grid-cols-4 gap-4 mt-6">
    <div class="skeleton h-28 rounded-2xl"></div>
    <div class="skeleton h-28 rounded-2xl"></div>
    <div class="skeleton h-28 rounded-2xl"></div>
    <div class="skeleton h-28 rounded-2xl"></div>
  </div>
  <!-- Resource usage skeleton -->
  <div class="grid grid-cols-3 gap-4 mt-6">
    <div class="skeleton h-36 rounded-2xl"></div>
    <div class="skeleton h-36 rounded-2xl"></div>
    <div class="skeleton h-36 rounded-2xl"></div>
  </div>
</template>
```

**Table/list page** (header + search + table rows):
```html
<template v-if="loading">
  <div class="skeleton h-28 rounded-2xl"></div>
  <div class="skeleton h-12 w-full max-w-md rounded-xl mt-6"></div>
  <div class="card bg-base-100/80 border border-base-300/20 shadow-sm mt-6 overflow-hidden">
    <table class="table w-full">
      <thead>
        <tr class="bg-base-200/50">
          <th v-for="i in 6" :key="i"><div class="skeleton h-4 w-16 rounded"></div></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="i in 5" :key="i">
          <td v-for="j in 6" :key="j">
            <div class="skeleton h-4 rounded" :class="j === 2 ? 'w-20' : j === 3 ? 'w-36' : 'w-16'"></div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
```

**Form page** (header + tabs + form cards + save button):
```html
<template v-if="loading">
  <div class="skeleton h-28 rounded-2xl"></div>
  <div class="flex gap-2 p-1.5 mt-6">
    <div class="skeleton h-10 w-32 rounded-xl"></div>
    <div class="skeleton h-10 w-32 rounded-xl"></div>
  </div>
  <div class="grid gap-5 sm:grid-cols-2 mt-6">
    <div class="skeleton h-44 rounded-2xl"></div>
    <div class="skeleton h-44 rounded-2xl"></div>
    <div class="skeleton h-44 rounded-2xl sm:col-span-2"></div>
  </div>
  <div class="flex justify-end pt-2 mt-6">
    <div class="skeleton h-12 w-36 rounded-xl"></div>
  </div>
</template>
```

**List item skeleton** (sessions, comments, etc.):
```html
<div v-for="i in 4" :key="i" class="flex items-center gap-3 p-3 rounded-xl bg-base-200/50">
  <div class="skeleton w-10 h-10 rounded-lg shrink-0"></div>
  <div class="flex-1 space-y-2">
    <div class="skeleton h-4 w-3/5 rounded"></div>
    <div class="flex gap-3">
      <div class="skeleton h-3 w-20 rounded"></div>
      <div class="skeleton h-3 w-24 rounded"></div>
      <div class="skeleton h-3 w-16 rounded"></div>
    </div>
  </div>
  <div class="skeleton h-8 w-8 rounded-lg shrink-0"></div>
</div>
```

### SSR safety with skeletons
- Skeletons are plain HTML `<div>` elements with CSS classes — they are **100% SSR-safe** and will never cause hydration issues.
- No browser APIs (`window`, `document`, etc.) are needed for skeleton rendering.

### Loading guard rule
- When a page component uses a loading pattern (`loading` ref starts as `true`, set to `false` in `onMounted`), every early return path in `onMounted` MUST set `loading.value = false` before returning. Failure to do so will leave the skeleton showing forever.
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

## API service usage rule (CRITICAL — do NOT use raw fetch for authenticated requests)
- **NEVER use bare `fetch()`** to call any backend API endpoint that requires authentication (admin routes, user routes, any endpoint protected by `AdminAuth` or `Auth` middleware). Always use the pre-built API services from `web/src/services/api.js`:
  - `adminAPI.get()`, `adminAPI.post()`, `adminAPI.put()`, `adminAPI.delete()` — for admin-authenticated requests. Automatically injects `Authorization: Bearer <token>`, handles token refresh, and redirects to login on 401.
  - `userAPI.get()`, `userAPI.post()`, etc. — for user-authenticated requests.
- Using raw `fetch()` without the `Authorization` header will cause **401 Unauthorized** errors because the backend middleware rejects requests without a valid JWT.
- The ONLY exception: **public endpoints** that do NOT require authentication (e.g., `/api/v1/system/info`, `/api/v1/oauth/providers`, login/register endpoints). For these, raw `fetch()` is acceptable.
- **Every new page or component that calls an authenticated API MUST import and use `adminAPI` or `userAPI`** — this is non-negotiable. Review all `fetch()` calls in new code to ensure they go through the proper API service when the endpoint requires auth.
- Example (correct):
  ```js
  import { adminAPI } from '../../services/api'
  // ...
  const resp = await adminAPI.get(`/api/v1/${adminPrefix}/oauth/settings`)
  const json = await resp.json()
  ```
- Example (WRONG — will cause 401):
  ```js
  // NEVER do this for authenticated endpoints:
  const resp = await fetch(`/api/v1/${adminPrefix}/oauth/settings`)
  ```

## Icon library
- **NEVER** use emoji for icons anywhere in the frontend
- Primary icon library: `lucide-vue-next` (Lucide icons)
- Secondary icon library: `@tabler/icons-vue` (Tabler Icons) — use when Lucide doesn't have the desired icon
- Install: `bun add lucide-vue-next @tabler/icons-vue`

- Usage example (Lucide):
  ```vue
  <script setup>
  import { Settings, User, LogOut } from 'lucide-vue-next'
  </script>
  <template>
    <Settings :size="20" />
  </template>
  ```

- Usage example (Tabler):
  ```vue
  <script setup>
  import { IconBrandGithub, IconBrandGoogle } from '@tabler/icons-vue'
  </script>
  <template>
    <IconBrandGithub :size="20" />
  </template>
  ```

## Language / i18n
- Backend loads translations from `lang/` directory (JSON files, nested structure)
- Frontend i18n module loads translations via `/lang/{locale}.json` (proxied from `web/public/lang/`)
- Cookie `purecore-locale` persists user language choice across requests
- Both sides flatten nested JSON to `"group.key"` format

## OAuth / Third-Party Login (Strategy Pattern)
- OAuth providers implement the `core/oauth.Provider` interface defined in `core/oauth/base.go`
- **Adding a new provider requires ONLY one new file**: `core/oauth/{provider}.go` that calls `oauth.Register()` in `init()`. No changes to routes, controllers, or frontend needed.
- Global registry at `core/oauth/registry.go` — providers self-register, admin settings page and login/register buttons auto-detect them
- State tokens use HMAC-SHA256 signing with 10-minute expiry (CSRF protection) in `core/oauth/state.go`
- Database: `oauth_accounts` table maps `(provider, provider_user_id)` → `user_id` via `app/Models/OAuthAccount.go`
- Backend controller at `app/Http/Controllers/OAuthController.go` handles: redirect, callback, bind, unbind, list accounts
- Routes defined in `routes/oauth.go`, registered from `cmd/serve.go`
- Frontend:
  - `web/src/composables/useOAuth.js` — fetches available providers, builds redirect URLs
  - `web/src/components/auth/OAuthButton.vue` — reusable button component for any provider
  - `web/src/pages/auth/OAuthCallback.vue` — callback handler with bind/register UX
  - `web/src/pages/admin/AdminOAuthSettings.vue` — tabbed admin panel to configure each provider (enabled, login/register toggles, client_id, client_secret, redirect_url)
- Admin settings keys follow pattern: `oauth_{provider}_enabled`, `oauth_{provider}_login_enabled`, `oauth_{provider}_register_enabled`, `oauth_{provider}_client_id`, `oauth_{provider}_client_secret`, `oauth_{provider}_redirect_url`
- Bind flow: when OAuth user doesn't exist, show two options: (1) register with pre-filled info, (2) login first then bind
- Documentation: `docs/en/OAUTH.md` and `docs/zh/OAUTH.md`
- Translation keys under `admin.oauth_*` (admin panel) and `oauth.*` (public-facing auth pages)

## Admin settings sidebar entry rule
- When creating a new admin settings page (or any new page under the admin route prefix), after writing the page component, adding its route in `web/src/router/routes.js`, and completing translations, **ALWAYS ask the user** whether they want to add a sidebar link in `web/src/components/AdminLayout.vue`. Do NOT automatically add sidebar entries without explicit user confirmation.
- The sidebar is managed in `web/src/components/AdminLayout.vue` and must be updated in both the desktop sidebar (`.hidden.lg:flex`) and mobile sidebar (`@click="closeSidebar"` sections) for consistency.
