# 服务端渲染 (SSR)

PureCore 内置基于 **Vue 3** + **Bun** + **Vite** 的 SSR 支持。SSR 服务器运行在端口 `9001`，在服务器端预渲染 Vue 应用，从而实现更快的初始页面加载和更好的 SEO。

## 架构

```
浏览器请求 → Bun SSR 服务器 (端口 9001) → Vite 开发服务器 (HMR) → Vue SSR 渲染 → HTML 响应
                                                   ↓
                                          服务端入口 (renderToString)
                                                   ↓
                                          客户端入口 (hydration)
```

## 关键文件

| 文件 | 用途 |
|------|---------|
| `web/server.js` | SSR 服务器，使用 `Bun.serve()`（生产模式）或 `node:http` + Vite 中间件（开发模式） |
| `web/src/entry-server.js` | 服务端入口点 — 使用 `createSSRApp` 创建 Vue 应用，渲染为 HTML |
| `web/src/entry-client.js` | 客户端入口点 — 对服务端渲染的 HTML 进行 hydration |
| `web/src/router/routes.js` | 共享路由定义，客户端和服务端共同使用 |
| `web/vite.config.js` | Vite 配置 — 使用 `VITE_API_PROTOCOL`、`VITE_API_HOST`、`VITE_API_PORT` 将 `/api` 代理到后端 |

## 环境配置

SSR 服务器从项目根目录的 `.env` 文件读取配置。SSR 的关键变量：

| 变量 | 描述 | 默认值 |
|----------|-------------|---------|
| `FRONTEND_PORT` / `SSR_PORT` | SSR 服务器监听端口 | `9001` |
| `VITE_API_PROTOCOL` | 后端 API 协议（`http`/`https`） | `http` |
| `VITE_API_HOST` | 后端 API 主机名 | `localhost` |
| `VITE_API_PORT` | 后端 API 端口 | `9002` |
| `THEME` | SSR 的默认 DaisyUI 主题 | `sunset` |
| `VITE_ADMIN_ROUTE_PREFIX` | 后台路由前缀 | `control-panel` |

### API 代理

在开发和生产模式下，SSR 服务器都将 `/api/*` 请求代理到后端，使用配置的协议、主机和端口：

- **开发模式**：使用 Node.js 的 `http`/`https` 模块转发请求
- **生产模式**：使用 Bun 内置的 `fetch()` API

这使得前端可以在一个端口上提供服务，同时 API 请求透明地转发到后端，避免 CORS 问题并简化部署。

### 主题处理

SSR 服务器读取 `THEME` 环境变量（或 `theme.config.json` 回退）来设置 `<html>` 元素上的初始 `data-theme` 属性。解析后的主题存储在 `purecore-theme` cookie 中，以便客户端在 hydration 时不会出现样式闪烁。优先级：

1. 环境变量 `THEME`
2. `web/theme.config.json` → `theme` 字段
3. 硬编码默认值（`'sunset'`）

## 启动 SSR 服务器

**开发模式**（带热重载）：
```bash
cd web
bun run dev
# → http://localhost:9001
```

**生产模式**（构建后）：
```bash
cd web
bun run build    # 构建客户端和服务端包
bun run preview  # 启动生产 SSR 服务器
```

## 工作原理

### 服务端

1. SSR 服务器（`web/server.js`）读取 `web/package.json` 获取项目元数据（`purecore` 字段），并读取 `lang/` 获取翻译文件
2. 每个请求调用 `entry-server.js` 中的 `render()` 函数
3. `entry-server.js` 使用 `createSSRApp` 创建 Vue SSR 应用，使用 `createMemoryHistory` 处理路由，并通过 `renderToString` 渲染为 HTML 字符串
4. HTML 被注入到模板的 `<!--ssr-outlet-->` 占位符中
5. 同步 CSS `<link>` 标签被添加到 `<head>` 中，防止 FOUC（无样式内容闪烁）

### 客户端

1. 浏览器接收完整渲染的 HTML
2. `entry-client.js` 使用 `createSSRApp` 和 `createWebHistory` 创建 Vue SSR 应用
3. Vue 对现有 DOM 进行 hydration，不重新渲染
4. 带有 `onMounted` 钩子的组件（如 `HomePage.vue` 获取版本信息）在 hydration 后运行

## 语言支持

SSR 服务器从 `Accept-Language` 请求头检测用户语言，并将其传递给渲染函数以及预加载的翻译文件。这使得服务器能够在正确的语言下渲染页面，而无需等待客户端 JavaScript。

```javascript
// 在 web/server.js 中
const acceptLang = req.headers['accept-language'] || ''
const locale = acceptLang.startsWith('zh') ? 'zh' : 'en'
await render(url, { locale, translations, projectInfo })
```

## 项目元数据注入

启动时读取 `web/package.json`（`purecore` 键下）中的项目元数据，并通过 `app.provide()` 注入到 SSR 上下文中。这使得 `HomePage.vue` 等组件无需发起 API 请求即可显示正确的版本号，消除了 hydration 不匹配。

```javascript
// 在 entry-server.js 中
if (projectInfo) {
  app.provide('projectInfo', projectInfo)
}

// 在 HomePage.vue 中
const ssrProjectInfo = inject('projectInfo', null)
```

## 为新页面添加 SSR 支持

1. 将页面组件添加到 `web/src/pages/`
2. 将路由添加到 `web/src/router/routes.js`
3. 确保组件正确处理 SSR：
   - 在初始渲染时避免访问 `window` 或 `document`（使用 `onMounted` 处理仅限浏览器的代码）
   - 使用 `inject()` 获取服务器提供的数据，而不是在挂载时使用 `fetch()`
   - 仅在客户端入口导入 `style.css`，不要在服务端入口导入

## 开发 vs 生产

### 开发模式

- 使用 Vite 的中间件模式进行热重载
- 通过 `node:http`/`node:https` 将 `/api` 代理到后端
- 启动时从项目根目录读取 `.env`

### 生产模式

- 使用 Bun 的 `Bun.serve()` 实现高性能 HTTP
- 从 `web/dist/client` 提供预构建的客户端资源
- 通过 Bun 的 `fetch()` 将 `/api` 代理到后端
- Docker Compose 部署时使用 `VITE_API_HOST=backend` 构建
