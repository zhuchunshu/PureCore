# OAuth / 第三方登录

PureCore 提供了一个生产就绪、可扩展的 OAuth 框架，用于添加第三方登录提供商（如 GitHub、Google 等）。架构遵循 **策略模式** — 每个提供商实现通用的 `Provider` 接口，新增平台时无需修改现有代码。

## 架构概览

| 层 | 文件 | 角色 |
|---|---|------|
| **接口** | `core/oauth/base.go` | 定义 `Provider` 契约（授权、交换令牌、获取用户信息） |
| **注册中心** | `core/oauth/registry.go` | 全局提供商注册中心；提供商通过 `init()` 自注册 |
| **基础实现** | `core/oauth/provider.go` | 可供新提供商嵌入的部分实现 |
| **状态管理** | `core/oauth/state.go` | 防 CSRF 的安全状态令牌（HMAC-SHA256 签名，10 分钟过期） |
| **数据库模型** | `app/Models/OAuthAccount.go` | 映射 OAuth 提供商 + 外部用户 ID → PureCore 用户 ID |
| **数据迁移** | `database/migrations/2026_05_03_000000_create_oauth_accounts_table.go` | `oauth_accounts` 表结构 |
| **控制器** | `app/Http/Controllers/OAuthController.go` | 重定向（启动 OAuth 流程）、回调（处理返回）、绑定/解绑接口 |
| **路由** | `routes/oauth.go` | 注册所有 OAuth 接口及其中间件 |
| **前端组合式函数** | `web/src/composables/useOAuth.js` | 获取可用提供商并构建重定向 URL |
| **前端组件** | `web/src/components/auth/OAuthButton.vue` | 适用于任何注册提供商的复用按钮 |
| **回调页面** | `web/src/pages/auth/OAuthCallback.vue` | 处理浏览器中的 OAuth 回调 |
| **管理员设置** | `web/src/pages/admin/AdminOAuthSettings.vue` | 管理员面板，用于启用/禁用提供商并配置凭据 |

## Provider 接口

每个 OAuth 提供商都必须实现 `core/oauth/base.go` 中定义的 `Provider` 接口：

```go
type Provider interface {
    // Name 返回提供商标识符（如 "github"、"google"）
    Name() string

    // DisplayName 返回在按钮和设置中显示的人类可读名称
    DisplayName() string

    // GetAuthURL 生成包含 state 和 redirect 的授权 URL
    GetAuthURL(state string, redirectURL string) string

    // ExchangeCode 使用授权码交换访问令牌
    ExchangeCode(code string, redirectURL string) (*Token, error)

    // GetUserInfo 使用访问令牌获取用户的个人资料
    GetUserInfo(token *Token) (*UserInfo, error)
}

type Token struct {
    AccessToken  string
    TokenType    string
    RefreshToken string // 可选
    ExpiresIn    int
}

type UserInfo struct {
    ProviderUserID string                 // 提供商端的唯一 ID
    Email          string
    Name           string
    AvatarURL      string
    RawData        map[string]interface{} // 原始响应，供自定义使用
}
```

`Provider` 接口是新提供商需要实现的唯一约定。`core/oauth/provider.go` 中提供了部分基础实现，可嵌入以减少常见 OAuth 2.0 流程的样板代码。

## 注册中心模式

全局注册中心（`core/oauth/registry.go`）允许提供商自行注册，无需修改任何现有代码：

```go
var registry = make(map[string]Provider)

func Register(p Provider) {
    registry[p.Name()] = p
}

func Get(name string) (Provider, bool) {
    p, ok := registry[name]
    return p, ok
}

func All() map[string]Provider {
    return registry
}
```

提供商通常在 `init()` 函数内调用 `Register()`，因此在程序启动时自动可用。

## 添加新提供商（以 GitHub 为例）

要添加 GitHub 登录，只需创建一个新文件：

### 第一步：创建提供商实现

创建 `core/oauth/github.go`：

```go
package oauth

import (
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
)

type GitHubProvider struct {
    *BaseProvider
}

func init() {
    Register(&GitHubProvider{})
}

func (p *GitHubProvider) Name() string { return "github" }
func (p *GitHubProvider) DisplayName() string { return "GitHub" }

func (p *GitHubProvider) GetAuthURL(state string, redirectURL string) string {
    u := "https://github.com/login/oauth/authorize"
    return u + "?client_id=" + p.getSetting("client_id") +
        "&redirect_uri=" + url.QueryEscape(redirectURL) +
        "&scope=user:email&state=" + state
}

func (p *GitHubProvider) ExchangeCode(code string, redirectURL string) (*Token, error) {
    // POST https://github.com/login/oauth/access_token
    // 携带 client_id, client_secret, code, redirect_uri
    // 解析响应获取 access_token
}

func (p *GitHubProvider) GetUserInfo(token *Token) (*UserInfo, error) {
    // GET https://api.github.com/user 带 Authorization: Bearer token
    // 解析 JSON 填充 UserInfo
}
```

### 第二步：注册设置项

管理面板期望每个提供商的配置以网站选项的形式存储。按照约定，键名遵循以下模式：

- `oauth_{provider}_enabled` — "1" 或 "0"
- `oauth_{provider}_login_enabled` — "1" 或 "0"
- `oauth_{provider}_register_enabled` — "1" 或 "0"
- `oauth_{provider}_client_id`
- `oauth_{provider}_client_secret`
- `oauth_{provider}_redirect_url`

`BaseProvider` 辅助类（`core/oauth/provider.go`）提供了 `getSetting(key)` 方法，可从选项存储中读取这些值。

### 第三步：运行数据迁移（如需额外数据表）

大多数提供商使用现有的 `oauth_accounts` 表即可正常工作。如果提供商需要额外数据，请创建新的迁移文件。

### 完成！

添加提供商文件后，它将自动可用 — 无需修改路由、控制器或前端代码。管理设置页面和登录/注册按钮将自动检测并显示它。

## OAuth 流程详解

1. **用户点击"使用 X 登录"** 按钮（位于登录或注册页面）。
   - `OAuthButton.vue` 使用 `useOAuth.js` 构建 URL，如 `/api/v1/auth/oauth/github/redirect?redirect=/dashboard`。
   - `redirect` 查询参数指定登录成功后的跳转目标。

2. **重定向接口**（`GET /api/v1/auth/oauth/{provider}/redirect`）：
   - 从注册中心查找提供商。
   - 生成签名的状态令牌（防 CSRF）。
   - 将 `redirect` 目标存储在状态令牌中。
   - 将浏览器重定向到提供商的授权页面。

3. **用户在提供商网站上授权**。

4. **提供商重定向回** 配置的回调 URL（`GET /api/v1/oauth/callback?code=...&state=...`）。
   - `OAuthCallback.vue` 提取 `code` 和 `state` 参数。
   - 调用后端接口 `POST /api/v1/auth/oauth/callback`，传入 `code` 和 `state`。

5. **后端回调接口**（`POST /api/v1/auth/oauth/callback`）：
   - 验证状态令牌（CSRF 检查、过期检查）。
   - 通过 `provider.ExchangeCode()` 使用授权码交换访问令牌。
   - 通过 `provider.GetUserInfo()` 获取用户信息。
   - 检查此提供商 + 提供商用户 ID 是否存在 `oauth_account` 记录：
     - **如果找到**：为关联用户签发 JWT + 刷新令牌，返回 `{ action: "logged_in", token, refresh_token }`。
     - **如果未找到**：返回 `{ action: "bind", oauth_id, provider, user_info: { email, name, avatar_url } }`。

6. **前端处理"绑定"情况**（在 `OAuthCallback.vue` 中）：
   - 显示两个选项：
     1. **使用提供商信息注册**：用提供商的邮箱和姓名预填充注册表单。注册成功后，后端会自动关联 OAuth 账户。
     2. **登录并绑定**：用户使用已有凭据登录。登录后关联 OAuth 账户。

7. **绑定接口**（`POST /api/v1/auth/oauth/bind`）：
   - 要求用户已认证（需要 JWT）。
   - 将 OAuth 账户关联到当前登录用户。

## 管理设置

管理员可以通过 **OAuth 设置页面**（`/{admin_prefix}/oauth`）配置每个 OAuth 提供商。该页面提供：

- **标签页界面**：每个注册的提供商显示为一个选项卡。
- **功能开关**：启用/禁用提供商，并单独控制是否在登录页面和/或注册页面显示。
- **凭据**：Client ID、Client Secret 和回调 URL 字段。
- 所有设置以网站选项的形式存储（通过 `app/Models/WebOption`），并自动被认证系统尊重。

## API 接口

| 方法 | 路径 | 认证 | 描述 |
|--------|------|------|-------------|
| `GET` | `/api/v1/auth/oauth/{provider}/redirect` | 无 | 重定向用户到提供商的授权页面 |
| `GET` | `/api/v1/oauth/callback` | 无 | 提供商授权后的回调 URL（页面） |
| `POST` | `/api/v1/auth/oauth/callback` | 无 | 交换授权码并处理登录/绑定 |
| `POST` | `/api/v1/auth/oauth/bind` | 用户 JWT | 将 OAuth 账户关联到已认证用户 |
| `DELETE` | `/api/v1/auth/oauth/unbind/{provider}` | 用户 JWT | 解除 OAuth 账户关联 |
| `GET` | `/api/v1/auth/oauth/accounts` | 用户 JWT | 列出用户的已关联 OAuth 账户 |
| `GET` | `/api/v1/auth/oauth/providers` | 无 | 列出可用 OAuth 提供商（用于前端按钮） |
| `GET` | `/api/v1/{admin_prefix}/oauth/settings` | 管理员 JWT | 获取管理员 OAuth 提供商设置 |
| `POST` | `/api/v1/{admin_prefix}/oauth/settings` | 管理员 JWT | 更新 OAuth 提供商设置 |

## 安全考量

1. **CSRF 防护**：状态令牌使用 HMAC-SHA256 签名，10 分钟后过期。后端在每次回调时验证状态。
2. **设置隔离**：OAuth 凭据（Client ID、Secret）存储在数据库中，仅管理员可访问。
3. **用户隐私**：`OAuthAccount` 模型仅存储提供商名称和外部用户 ID — 原始 OAuth 用户数据不会持久化。
4. **令牌安全**：OAuth 访问令牌永远不会存储在服务器端。它们仅在回调期间用于获取用户信息，随后即被丢弃。
5. **管理员专属配置**：仅已认证的管理员可以查看或修改 OAuth 提供商设置。
