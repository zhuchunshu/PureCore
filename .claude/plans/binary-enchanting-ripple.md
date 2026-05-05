
# OAuth 未关联回调 — 独立登录/注册页面实现计划

## Context

当前 OAuth 授权回调后，如果 OAuth 账户未关联本站用户，`OAuthCallback.vue` 内嵌了两个处理模式：
1. **注册模式** (`mode='register'`)：一个简陋的 name+email 表单，调用 `POST /api/v1/oauth/register`
2. **登录后绑定模式** (`mode='bind'`)：重定向到 `/login` 页面，登录后再跳回 callback

**问题**：注册表单过于简陋（无密码、无 Turnstile），且两种模式都耦合在 `OAuthCallback.vue` 内部，不符合项目"独立页面/独立 controller"的架构风格。

**目标**：创建独立的后端 Controller、前端 Page、路由，分别处理 OAuth 未关联后的「注册 + 绑定」和「登录 + 绑定」两个流程。新页面**视觉和交互逻辑**与现有 `LoginPage.vue` / `RegisterPage.vue` 完全一致。

---

## 实施步骤

### 1. 后端 — 新建 OAuthLinkController

**文件**: `app/Http/Controllers/OAuthLinkController.go` (新文件)

包含两个方法：

#### 1.1 `LinkRegister` — 注册新用户 + 绑定 OAuth
- **路由**: `POST /api/v1/oauth/:provider/link/register`
- **请求体**: `{ link_token, name, email, password, turnstile_token }`
- **流程**:
  1. 验证 link_token（调用 `oauth.ParseLinkToken()`）
  2. 验证 Turnstile（同 `UserAuthController.Register`）
  3. 检查 email 是否已存在（存在返回 409）
  4. 创建 `models.User`（用真实密码，而非随机占位符）
  5. 创建 `models.OAuthAccount` 关联记录
  6. 生成 JWT token，返回 `{ token, refresh_token, name, email }`
- **复用**: `OAuthController.loginUser()` 的 token 生成/返回逻辑

#### 1.2 `LinkLogin` — 登录已有用户 + 绑定 OAuth
- **路由**: `POST /api/v1/oauth/:provider/link/login`
- **请求体**: `{ link_token, email, password, turnstile_token }`
- **流程**:
  1. 验证 link_token
  2. 验证 Turnstile
  3. 验证 email/password（同 `UserAuthController.Login`）
  4. 检查该 OAuth provider_id 是否已被其他用户绑定（若已绑定则返回 409）
  5. 绑定 OAuth 账户到当前用户（插入 `OAuthAccount` 记录）
  6. 若用户 avatar 为空，用 OAuth avatar 填充
  7. 生成 JWT token，返回 `{ token, refresh_token, name, email }`

### 2. 后端 — 注册路由

**文件**: `routes/oauth_link.go` (新文件)
```go
func RegisterOAuthLinkRoutes(router *core.Router) {
    ctrl := &controllers.OAuthLinkController{}
    router.Post("/api/v1/oauth/:provider/link/register", core.H(ctrl.LinkRegister))
    router.Post("/api/v1/oauth/:provider/link/login", core.H(ctrl.LinkLogin))
}
```

**文件**: `cmd/serve.go`
- 在 `Boot()` 之后调用 `routes.RegisterOAuthLinkRoutes(router)`

### 3. 前端 — 新建 OAuthLinkRegister 页面

**文件**: `web/src/pages/auth/OAuthLinkRegister.vue` (新文件)

**完全复用 `RegisterPage.vue` 的视觉和交互**：
- 同样的背景（animated grid + glow orbs）
- 同样的插画布局（left illustration + right form card）
- 同样的表单字段：name、email、password、confirmPassword
- 同样的输入框样式（SVG 图标 + 圆角输入框）
- 同样的 Turnstile widget
- 同样的提交按钮样式（gradient button）
- 同样的错误提示样式（alert）

**差异点**：
- name 和 email 从 URL query params 预填充（`route.query.name`, `route.query.email`）
- 提交调用 `POST /api/v1/oauth/:provider/link/register` 而非 `/api/v1/auth/register`
- 请求体额外包含 `link_token`
- 底部链接改为「已有账户？登录并绑定」→ 跳转到 `/oauth/:provider/link/login`
- 无 OAuth 第三方登录按钮（因为已经在 OAuth 流程中）
- SEO title 改为 "Complete Registration - {provider}"

### 4. 前端 — 新建 OAuthLinkLogin 页面

**文件**: `web/src/pages/auth/OAuthLinkLogin.vue` (新文件)

**完全复用 `LoginPage.vue` 的视觉和交互**：
- 同样的背景、插画、表单卡片布局
- 同样的 email + password 字段
- 同样的 Turnstile widget
- 同样的提交按钮样式
- 同样的错误提示样式

**差异点**：
- 提交调用 `POST /api/v1/oauth/:provider/link/login` 而非 `/api/v1/auth/login`
- 请求体额外包含 `link_token`
- 底部链接改为「没有账户？注册并绑定」→ 跳转到 `/oauth/:provider/link/register`
- 无 OAuth 第三方登录按钮
- SEO title 改为 "Login to Link - {provider}"

### 5. 前端 — 更新 OAuthCallback.vue

**修改**: `web/src/pages/auth/OAuthCallback.vue`

- `chooseRegister()` → 改为 `router.push()` 到 `/oauth/:provider/link/register?link_token=...&name=...&email=...&avatar_url=...&redirect=...`
- `chooseLogin()` → 改为 `router.push()` 到 `/oauth/:provider/link/login?link_token=...&email=...&redirect=...`
- **删除**内部的 `mode='register'` 表单部分（register form 卡片），因为已迁移到独立页面
- **删除** `mode='login'` 的 `sessionStorage` + redirect 逻辑，改为直接导航
- 保留 `choose` 模式（两个按钮卡片）

### 6. 前端 — 注册路由

**文件**: `web/src/router/routes.js`

新增两个路由：
```js
{
  path: '/oauth/:provider/link/register',
  name: 'OAuthLinkRegister',
  component: () => import('../pages/auth/OAuthLinkRegister.vue'),
  props: true,
},
{
  path: '/oauth/:provider/link/login',
  name: 'OAuthLinkLogin',
  component: () => import('../pages/auth/OAuthLinkLogin.vue'),
  props: true,
},
```

### 7. i18n 翻译补充

**文件**: `web/src/i18n.js` 或 `lang/en.json` / `lang/zh.json`

新增 key：
- `oauth.link_register_title` → "Complete Registration"
- `oauth.link_login_title` → "Login to Link"
- `oauth.link_register_desc` → "Create an account and link your {provider} account"
- `oauth.link_login_desc` → "Login to your existing account to link {provider}"
- `oauth.already_have_account_link` → "Already have an account? Login to link"
- `oauth.no_account_link` → "Don't have an account? Register to link"

---

## 涉及修改的文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `app/Http/Controllers/OAuthLinkController.go` | **新建** | 两个端点：LinkRegister、LinkLogin |
| `routes/oauth_link.go` | **新建** | 路由注册 |
| `cmd/serve.go` | **修改** | 调用 `RegisterOAuthLinkRoutes` |
| `web/src/pages/auth/OAuthLinkRegister.vue` | **新建** | 复制 RegisterPage 样式 + OAuth 绑定逻辑 |
| `web/src/pages/auth/OAuthLinkLogin.vue` | **新建** | 复制 LoginPage 样式 + OAuth 绑定逻辑 |
| `web/src/pages/auth/OAuthCallback.vue` | **修改** | chooseLogin/chooseRegister 改为页面跳转，删除内嵌表单 |
| `web/src/router/routes.js` | **修改** | 新增两条路由 |
| `web/src/composables/useOAuth.js` | **修改**（可选） | 新增 `linkLogin()` / `linkRegister()` 辅助方法 |
| `web/src/i18n.js` | **修改** | 新增 OAuth link 相关翻译 key |

---

## 验证方式

1. 启动后端：`./purecore serve`
2. 启动前端：`cd web && bun run dev`
3. 配置一个 OAuth provider（如 GitHub），确保 `login_enabled` 和 `register_enabled` 为 true
4. 在无痕窗口中访问 `/login`，点击 GitHub OAuth 按钮
5. 授权后应被重定向到 OAuthCallback 页面，显示两个选项卡片
6. 点击「Register with GitHub」→ 跳转到 OAuthLinkRegister 页面（样式与 RegisterPage 一致）
7. 填写密码并提交 → 应成功创建账户、绑定 GitHub、自动登录、跳转到目标页面
8. 再次点击 GitHub OAuth 按钮 → 用另一个 GitHub 账号授权
9. 在 OAuthCallback 点击「Login to link」→ 跳转到 OAuthLinkLogin 页面（样式与 LoginPage 一致）
10. 用步骤 7 的账户登录 → 应成功绑定第二个 GitHub 账号、自动跳转
11. 验证 Integrations 页面（`/dashboard/integrations`）显示两个已绑定的 GitHub 账户
