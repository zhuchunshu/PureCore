# Server-Side Rendering (SSR)

PureCore includes built-in SSR support using **Vue 3** + **Bun** + **Vite**. The SSR server runs on port `9001` and pre-renders the Vue application on the server for faster initial page loads and better SEO.

## Architecture

```
Browser Request → Bun SSR Server (port 9001) → Vite Dev Server (HMR) → Vue SSR Render → HTML Response
                                                                              ↓
                                                                    Entry Server (renderToString)
                                                                              ↓
                                                                    Entry Client (hydration)
```

## Key Files

| File | Purpose |
|------|---------|
| `web/server.js` | SSR server using `Bun.serve()` (production) or `node:http` + Vite middleware (development) |
| `web/src/entry-server.js` | Server-side entry point — creates the Vue app with `createSSRApp`, renders to HTML |
| `web/src/entry-client.js` | Client-side entry point — hydrates the server-rendered HTML |
| `web/src/router/routes.js` | Shared route definitions used by both server and client |

## Starting the SSR Server

**Development mode** (with HMR):
```bash
cd web
bun run dev
# → http://localhost:9001
```

**Production mode** (after building):
```bash
cd web
bun run build    # Builds both client and server bundles
bun run preview  # Starts production SSR server
```

## How It Works

### Server-Side

1. The SSR server (`web/server.js`) reads `web/package.json` for project metadata (`purecore` field) and `lang/` for translation files
2. On each request, it calls `render()` from `entry-server.js`
3. `entry-server.js` creates a Vue SSR app with `createSSRApp`, uses `createMemoryHistory` for routing, and renders to HTML string using `renderToString`
4. The HTML is injected into the template at `<!--ssr-outlet-->` placeholder
5. A synchronous CSS `<link>` tag is added to the `<head>` to prevent FOUC (Flash of Unstyled Content)

### Client-Side

1. The browser receives the fully rendered HTML
2. `entry-client.js` creates a Vue SSR app with `createSSRApp` and `createWebHistory`
3. Vue hydrates the existing DOM without re-rendering
4. Components with `onMounted` hooks (like `HomePage.vue` fetching version info) run after hydration

## Language Support

The SSR server detects the user's language from the `Accept-Language` header and passes it to the render function along with pre-loaded translation files. This allows the server to render the page in the correct language without waiting for client-side JavaScript.

```javascript
// In web/server.js
const acceptLang = req.headers['accept-language'] || ''
const locale = acceptLang.startsWith('zh') ? 'zh' : 'en'
await render(url, { locale, translations, projectInfo })
```

## Project Metadata Injection

Project metadata from `web/package.json` (under the `purecore` key) is read at startup and injected into the SSR context via `app.provide()`. This allows components like `HomePage.vue` to display the correct version number without making an API call, eliminating hydration mismatches.

```javascript
// In entry-server.js
if (projectInfo) {
  app.provide('projectInfo', projectInfo)
}

// In HomePage.vue
const ssrProjectInfo = inject('projectInfo', null)
```

## Adding SSR Support to New Pages

1. Add your page component to `web/src/pages/`
2. Add the route to `web/src/router/routes.js`
3. Ensure your component handles SSR by:
   - Avoiding `window` or `document` access during initial render (use `onMounted` for browser-only code)
   - Using `inject()` for any server-provided data instead of `fetch()` on mount
   - Importing `style.css` only in client entry, not server entry
