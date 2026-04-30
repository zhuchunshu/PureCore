# 认证系统

PureCore 实现了双 JWT 认证系统，为普通用户和管理员提供访问令牌和刷新令牌。

## 概述

| 用户类型 | 访问令牌 (JWT) | 刷新令牌 | 密码哈希 |
|----------|---------------|----------|----------|
| 普通用户 | 15 分钟 | 7 天，存储在数据库中 | bcrypt |
| 管理员 | 15 分钟 | 7 天，存储在数据库中 | bcrypt |

两种用户类型使用独立的 JWT 密钥和令牌生成函数，确保用户认证和管理员认证之间的完全隔离。

## 用户认证

### 注册

普通用户通过 `POST /api/v1/auth/register` 注册。系统会：

1. 验证请求体（`name`、`email`、`password`）
2. 可选地验证 Turnstile 令牌（如果启用了 `turnstile_public_login` 选项）
3. 检查邮箱地址是否重复
4. 使用 bcrypt 对密码进行哈希处理
5. 创建用户记录
6. 生成访问令牌（JWT，15 分钟）和刷新令牌（随机字符串，7 天）
7. 将刷新令牌保存到 `users` 表中

**User 模型字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| name | string | 显示名称 |
| email | string | 邮箱（唯一） |
| password | string | Bcrypt 哈希 |
| avatar | string | 头像 URL（可为空） |
| bio | string | 用户简介（可为空） |
| status | int | 账户状态 |
| email_verified_at | time | 邮箱验证时间戳（可为空） |
| last_login_at | time | 最后登录时间戳（可为空） |
| refresh_token | string | 当前刷新令牌 |
| created_at | time | 创建时间戳 |
| updated_at | time | 更新时间戳 |
| deleted_at | time | 软删除时间戳 |

### 登录

用户通过 `POST /api/v1/auth/login` 登录。系统会：

1. 验证凭据
2. 可选地验证 Turnstile
3. 通过邮箱查找用户
4. 验证密码哈希
5. 生成新的访问和刷新令牌
6. 更新数据库中的 `refresh_token` 和 `last_login_at`

### 令牌刷新

用户通过 `POST /api/v1/auth/refresh` 刷新访问令牌。系统会：

1. 验证提供的刷新令牌
2. 通过刷新令牌查找用户
3. 生成新的访问令牌
4. 旋转刷新令牌（生成新的，替换数据库中的旧令牌）以提升安全性

### 个人资料

已认证用户可以通过 `GET /api/v1/auth/profile` 访问个人资料。该端点返回所有个人资料字段，包括头像、简介、状态和时间戳。

## 管理员认证

### 管理员注册

管理员通过 `POST /{admin_prefix}/auth/register` 注册。系统会：

1. 验证请求体（`username`、`password`、`name`）
2. 可选地验证 Turnstile（如果启用了 `turnstile_admin_register` 选项）
3. 分配 `"admin"` 角色（始终）
4. 创建管理员记录
5. 返回访问和刷新令牌

**AdminUser 模型字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint | 主键 |
| username | string | 登录用户名（唯一） |
| name | string | 显示名称 |
| password | string | Bcrypt 哈希 |
| role | string | 角色：`"admin"` 或 `"super_admin"` |
| token_version | int | 修改密码时递增以作废令牌 |
| refresh_token | string | 当前刷新令牌 |
| created_at | time | 创建时间戳 |
| updated_at | time | 更新时间戳 |

### 管理员登录

管理员通过 `POST /{admin_prefix}/auth/login` 登录。流程与用户登录相同，但使用管理员专用的令牌生成函数（`GenerateAdminToken`）。

管理员 JWT 包含额外的声明：

```json
{
  "id": 1,
  "username": "admin",
  "role": "admin",
  "token_version": 0,
  "exp": 1234567890
}
```

`token_version` 声明由 `AdminAuth` 中间件检查。如果与数据库中的当前值不匹配，令牌将被拒绝——这实现了修改密码时的强制退出功能。

### 密码修改与令牌作废

管理员可以通过 `POST /{admin_prefix}/auth/change-password` 修改密码。该端点：

1. 验证当前密码和新密码
2. 验证当前密码是否匹配
3. 对新密码进行哈希处理
4. 增加 `token_version`（作废所有现有 JWT）
5. 清除 `refresh_token`（作废所有刷新令牌）

此操作后，管理员必须重新登录以获取新令牌。

### 管理员刷新令牌

管理员通过 `POST /{admin_prefix}/auth/refresh` 刷新访问令牌。流程与用户令牌刷新相同，但使用管理员专用的令牌生成函数。

## 中间件

### Auth（用户）

位于 `app/Http/Middleware/Auth.go`。该中间件：

1. 从 `Authorization` 请求头中提取 Bearer 令牌
2. 使用 `GenerateUserToken` 验证 JWT
3. 将用户声明存储在 `c.Locals("user")` 中，格式为 `map[string]string{"id": "..."}`
4. 如果令牌缺失、无效或过期，返回 401

### AdminAuth

位于 `app/Http/Middleware/AdminAuth.go`。该中间件：

1. 从 `Authorization` 请求头中提取 Bearer 令牌
2. 使用 `GenerateAdminToken` 验证 JWT
3. 检查 `token_version` 是否与数据库匹配——如果不匹配则拒绝
4. 将管理员声明存储在 `c.Locals("user")` 中
5. 如果任何检查失败，返回 401

### 命名中间件注册

两个中间件都在 `RouteServiceProvider` 中注册为命名中间件：

```go
router.RegisterNamedMiddlewares(map[string]core.MiddlewareFunc{
    "auth":       middleware.Auth(),
    "admin_auth": middleware.AdminAuth(),
})
```

## JWT 令牌生成

### 用户令牌

```go
// 在 app/Http/Middleware/Auth.go 中
func GenerateUserToken(userID uint, userName string) (string, error)
```

声明：`id`、`name`、`exp`（从现在起 15 分钟）。
使用环境变量中的 `JWT_SECRET`。

### 管理员令牌

```go
// 在 app/Http/Middleware/AdminAuth.go 中
func GenerateAdminToken(adminID uint, username string, tokenVersion int) (string, error)
```

声明：`id`、`username`、`role`、`token_version`、`exp`（从现在起 15 分钟）。
使用环境变量中的 `JWT_SECRET`。

### 刷新令牌

```go
func GenerateRefreshToken() (string, error)
```

生成一个随机的 64 字符十六进制字符串。有效期为 7 天（由客户端强制执行，不是加密级别）。

## 客户端认证

### 用户认证（useUserAuth.js）

`web/src/composables/useUserAuth.js` 组合式函数处理：

- 登录、注册和退出流程
- localStorage 中的令牌存储
- 自动令牌刷新
- 响应式用户状态

### 管理员认证（useAuth.js）

`web/src/composables/useAuth.js` 组合式函数处理：

- 管理员登录、注册和退出流程
- localStorage 中的令牌存储
- 过期前自动刷新令牌
- 响应式管理员用户状态

## 安全注意事项

1. **密码哈希**：所有密码在存储前都使用 bcrypt 进行哈希处理
2. **令牌轮换**：每次使用刷新令牌时都会轮换——旧令牌立即失效
3. **令牌版本控制**：管理员 JWT 包含 `token_version` 声明——修改密码会使所有设备上的所有令牌失效
4. **短生命周期访问令牌**：15 分钟的过期时间最小化了令牌被盗用的时间窗口
5. **唯一密钥**：在生产环境中使用强唯一 `JWT_SECRET`（切勿使用默认值）
6. **Turnstile 保护**：可选择对登录/注册端点要求 Cloudflare Turnstile 验证
