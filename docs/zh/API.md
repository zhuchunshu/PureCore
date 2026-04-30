# PureCore API 接口文档

## 基础信息

- 基础路径: `/api/v1`
- 后台基础路径: `/{admin_prefix}`（默认：`/control-panel`，可通过 `ADMIN_ROUTE_PREFIX` 配置）
- 默认端口: `9002`
- 认证方式: `Bearer Token`（JWT）
- 响应格式: JSON

### 统一响应结构

所有接口返回统一的 JSON 格式：

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| code | int | 状态码，0 表示成功，非 0 表示错误 |
| message | string | 提示信息，支持中英文 |
| data | object/array | 返回数据，可能为 null |

### 分页响应

```json
{
  "code": 0,
  "message": "操作成功",
  "data": [],
  "total": 100,
  "page": 1,
  "per_page": 15
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| total | int | 总记录数 |
| page | int | 当前页码 |
| per_page | int | 每页条数 |

### 错误状态码

| HTTP 状态码 | code | 说明 |
|-------------|------|------|
| 200 | 0 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未授权 |
| 404 | 404 | 资源不存在 |
| 409 | 409 | 冲突（如邮箱重复） |
| 422 | 422 | 数据验证失败 |
| 500 | 500 | 服务器内部错误 |

## 公开接口

### 健康检查

```
GET /api/v1/ping
```

**请求示例**

```bash
curl http://localhost:9002/api/v1/ping
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": "pong"
}
```

### 项目信息

```
GET /api/v1/system/info
```

从 `purecore.json` 返回项目元数据，包括版本号、发布类型、作者、许可证和依赖项信息。

**请求示例**

```bash
curl http://localhost:9002/api/v1/system/info
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "name": "PureCore",
    "version": "1.0.0",
    "release_type": "alpha",
    "author": {
      "name": "zhuchunshu",
      "email": "",
      "url": "https://github.com/zhuchunshu"
    },
    "repository": {
      "type": "git",
      "url": "https://github.com/zhuchunshu/PureCore.git"
    },
    "license": "MIT",
    "keywords": ["go", "gofiber", "vue", "laravel-style", "framework"],
    "go_version": "1.21",
    "dependencies": {
      "backend": {
        "framework": "GoFiber v3",
        "database": "PostgreSQL",
        "orm": "GORM",
        "validation": "go-playground/validator"
      },
      "frontend": {
        "framework": "Vue 3",
        "build_tool": "Vite",
        "css": "Tailwind CSS + DaisyUI",
        "package_manager": "Bun"
      }
    },
    "description": {
      "en": "A full-stack Go web development framework wrapping GoFiber v3 in a Laravel-like development style.",
      "zh": "基于 Go 语言的全栈 Web 开发框架，将 GoFiber v3 封装成类似 Laravel 的开发风格。"
    }
  }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 项目名称 |
| version | string | 语义化版本号（如 1.0.0） |
| release_type | string | 发布阶段：`alpha`、`beta`、`rc` 或 `stable` |
| author | object | 作者信息（姓名、邮箱、网址） |
| license | string | 许可证标识符 |
| keywords | array | 项目标签 |
| description | object | 多语言描述（按语言代码索引） |

### 文档

```
GET /api/v1/docs?file=en/README.md
```

返回 `docs/` 目录下指定文档文件的内容。

**查询参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| file | string | 是 | 相对于 `docs/` 的文档文件路径（如 `en/README.md`） |

**请求示例**

```bash
curl "http://localhost:9002/api/v1/docs?file=zh/README.md"
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "filename": "zh/README.md",
    "content": "# PureCore 框架\n\n..."
  }
}
```

### 文档列表

```
GET /api/v1/docs/list
```

返回所有可用文档文件的列表。

**请求示例**

```bash
curl http://localhost:9002/api/v1/docs/list
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": ["en/README.md", "en/API.md", "zh/README.md", "zh/API.md"]
}
```

## 公开选项

```
GET /{admin_prefix}/options
```

返回所有公开的网站选项（站点名称、描述、关键词、Turnstile 站点密钥等）。此接口**不需要**认证。

**请求示例**

```bash
curl http://localhost:9002/control-panel/options
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "site_name": "PureCore",
    "site_description": "基于 Go 的全栈框架",
    "site_keywords": "go, vue, framework",
    "turnstile_site_key": "1x00000000000000000000AA"
  }
}
```

## 用户认证

### 注册

```
POST /api/v1/auth/register
```

创建新的普通用户账户。成功时返回 JWT 访问令牌和刷新令牌。

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 用户名，最少 2 个字符 |
| email | string | 是 | 邮箱，需符合邮箱格式且唯一 |
| password | string | 是 | 密码，最少 6 个字符 |
| turnstile_token | string | 条件 | Turnstile 令牌（当启用公网登录 Turnstile 时必填） |

**请求示例**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"password123"}' \
  http://localhost:9002/api/v1/auth/register
```

**响应示例（成功）**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "abc123def456...",
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

**响应示例（邮箱已存在）**

```json
{
  "code": 409,
  "message": "邮箱已存在"
}
```

### 登录

```
POST /api/v1/auth/login
```

验证普通用户身份并返回 JWT 令牌。更新 `last_login_at` 时间戳。

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 注册邮箱 |
| password | string | 是 | 密码，最少 6 个字符 |
| turnstile_token | string | 条件 | Turnstile 令牌（当启用公网登录 Turnstile 时必填） |

**请求示例**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}' \
  http://localhost:9002/api/v1/auth/login
```

**响应示例（成功）**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "abc123def456...",
    "name": "Alice",
    "email": "alice@example.com"
  }
}
```

**响应示例（凭据无效）**

```json
{
  "code": 401,
  "message": "用户名或密码错误"
}
```

### 刷新令牌

```
POST /api/v1/auth/refresh
```

使用有效的刷新令牌生成新的访问令牌。出于安全考虑，旧的刷新令牌会被轮换（替换为新的）。

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| refresh_token | string | 是 | 登录/注册返回的有效刷新令牌 |

**请求示例**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"abc123def456..."}' \
  http://localhost:9002/api/v1/auth/refresh
```

**响应示例（成功）**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "new-refresh-token..."
  }
}
```

### 获取个人信息

```
GET /api/v1/auth/profile
```

返回当前认证用户的个人资料。需要在请求头中携带 `Authorization: Bearer <token>`。

**请求示例**

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:9002/api/v1/auth/profile
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "id": 1,
    "name": "Alice",
    "email": "alice@example.com",
    "avatar": null,
    "bio": null,
    "status": 1,
    "email_verified_at": null,
    "last_login_at": "2026-04-30T10:30:00Z",
    "created_at": "2026-04-28T08:00:00Z",
    "updated_at": "2026-04-30T10:30:00Z"
  }
}
```

## 管理员认证

> 所有管理员接口使用可配置的前缀（默认：`/control-panel`，通过 `ADMIN_ROUTE_PREFIX` 设置）。

### 检查管理员是否存在

```
GET /{admin_prefix}/auth/check
```

返回是否存在任何管理员用户。前端用于判断显示注册页面还是登录页面。

**请求示例**

```bash
curl http://localhost:9002/control-panel/auth/check
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "exists": true,
    "count": 1
  }
}
```

### 注册管理员

```
POST /{admin_prefix}/auth/register
```

创建新的管理员用户。始终分配 `"admin"` 角色。创建首个管理员无需事先认证。

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 管理员用户名，最少 3 个字符 |
| password | string | 是 | 密码，最少 6 个字符 |
| name | string | 是 | 显示名称 |
| turnstile_token | string | 条件 | Turnstile 令牌（当启用管理员注册 Turnstile 时必填） |

**请求示例**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123","name":"管理员"}' \
  http://localhost:9002/control-panel/auth/register
```

**响应示例（成功）**

```json
{
  "code": 0,
  "message": "管理员注册成功",
  "data": {
    "message": "管理员注册成功",
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "abc123def456...",
    "username": "admin",
    "name": "管理员",
    "role": "admin"
  }
}
```

### 管理员登录

```
POST /{admin_prefix}/auth/login
```

验证管理员身份并返回 JWT 令牌。访问令牌包含 `token_version`，支持强制退出（修改密码会增加 `token_version`，使所有现有令牌失效）。

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 管理员用户名，最少 3 个字符 |
| password | string | 是 | 密码，最少 6 个字符 |
| turnstile_token | string | 条件 | Turnstile 令牌（当启用管理员登录 Turnstile 时必填） |

**请求示例**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  http://localhost:9002/control-panel/auth/login
```

**响应示例（成功）**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "abc123def456...",
    "username": "admin",
    "name": "管理员",
    "role": "admin"
  }
}
```

### 管理员刷新令牌

```
POST /{admin_prefix}/auth/refresh
```

使用有效的刷新令牌生成新的访问令牌。旧的刷新令牌会被轮换。

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| refresh_token | string | 是 | 登录/注册返回的有效刷新令牌 |

**请求示例**

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"abc123def456..."}' \
  http://localhost:9002/control-panel/auth/refresh
```

**响应示例（成功）**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "new-refresh-token..."
  }
}
```

### 获取管理员信息

```
GET /{admin_prefix}/auth/profile
```

返回当前认证管理员的个人资料。需要管理员认证（`Authorization: Bearer <admin_token>`）。

**请求示例**

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:9002/control-panel/auth/profile
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "id": 1,
    "username": "admin",
    "name": "管理员",
    "role": "admin"
  }
}
```

### 修改管理员密码

```
POST /{admin_prefix}/auth/change-password
```

修改当前管理员的密码。**会增加 `token_version`**，使所有现有令牌（包括当前令牌）失效。客户端必须在调用后重新登录。

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| current_password | string | 是 | 当前密码，最少 6 个字符 |
| new_password | string | 是 | 新密码，最少 6 个字符 |

**请求示例**

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{"current_password":"admin123","new_password":"newpass456"}' \
  http://localhost:9002/control-panel/auth/change-password
```

**响应示例（成功）**

```json
{
  "code": 0,
  "message": "密码修改成功",
  "data": {
    "message": "密码修改成功"
  }
}
```

### 管理员选项（设置）

```
POST /{admin_prefix}/options
```

设置网站选项（键值对）。需要管理员认证。接受一个 JSON 对象，其中键为选项名称，值为选项值。

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| (动态键) | string | 否 | 任意选项键值对 |

**请求示例**

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{"site_name":"我的网站","site_description":"我的网站描述"}' \
  http://localhost:9002/control-panel/options
```

**响应示例（成功）**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "message": "选项保存成功"
  }
}
```

## 需认证的用户接口

> 以下所有接口需要在请求头携带 `Authorization: Bearer <token>`，使用有效的用户 JWT 令牌。

### 获取用户列表

```
GET /api/v1/users
```

**请求示例**

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:9002/api/v1/users
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": [
    { "id": 1, "name": "Alice", "email": "alice@example.com" },
    { "id": 2, "name": "Bob", "email": "bob@example.com" }
  ]
}
```

### 创建用户

```
POST /api/v1/users
```

**请求体**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 用户名，最少 2 个字符 |
| email | string | 是 | 邮箱，需符合邮箱格式 |

**请求示例**

```bash
curl -X POST \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  -H "Content-Type: application/json" \
  -d '{"name":"Charlie","email":"charlie@example.com"}' \
  http://localhost:9002/api/v1/users
```

**响应示例（成功）**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": {
    "name": "Charlie",
    "email": "charlie@example.com"
  }
}
```

### 获取用户详情

```
GET /api/v1/users/:id
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 用户 ID |

**请求示例**

```bash
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..." \
  http://localhost:9002/api/v1/users/1
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": { "id": 1, "name": "Alice", "email": "alice@example.com" }
}
```

**响应示例（未找到）**

```json
{
  "code": 404,
  "message": "用户不存在"
}
```

## 认证说明

### 令牌类型

| 令牌 | 有效期 | 用途 |
|------|------|------|
| 访问令牌 (JWT) | 15 分钟 | 用于认证 API 请求 |
| 刷新令牌 | 7 天 | 用于生成新的访问令牌 |

### 令牌流程

1. **注册**或**登录** → 服务器返回 `token`（访问令牌）+ `refresh_token`（刷新令牌）
2. 在所有需要认证的请求中使用 `Authorization: Bearer <token>` 请求头
3. 当 `token` 过期时 → 调用 `/auth/refresh` 并传入 `refresh_token` 获取新令牌对
4. **仅管理员**：修改密码会增加 `token_version`，使所有现有令牌失效

### 未认证响应

```json
{
  "code": 401,
  "message": "未授权"
}
```

## 跨域说明

后端已配置 CORS 中间件，允许以下跨域请求：

- 允许来源: 所有 (`*`)
- 允许方法: `GET`, `POST`, `PUT`, `DELETE`, `PATCH`
- 允许请求头: `Origin`, `Content-Type`, `Accept`, `Authorization`

## 多语言支持

后端支持根据 `Accept-Language` 请求头切换响应语言。

```bash
# 中文 (默认)
curl http://localhost:9002/api/v1/ping

# 英文
curl -H "Accept-Language: en" http://localhost:9002/api/v1/ping
```

中文环境下返回 `"message": "操作成功"`，英文环境下返回 `"message": "Operation successful"`。

## Turnstile 集成

部分接口在启用相应选项时可能需要 Cloudflare Turnstile 令牌。配置详情请参阅 [Turnstile 文档](./TURNSTILE.md)。
