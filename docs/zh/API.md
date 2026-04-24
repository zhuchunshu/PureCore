# PureCore API 接口文档

## 基础信息

- 基础路径: `/api/v1`
- 默认端口: `9002`
- 认证方式: `Bearer Token`
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

从 `web/package.json`（`purecore` 键下）返回项目元数据，包括版本号、发布类型、作者、许可证和依赖项信息。

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

## 认证接口（需登录）

> 所有以下接口需要在请求头携带 `Authorization: Bearer <token>`

### 获取用户列表

```
GET /api/v1/users
```

**请求头**

```
Authorization: Bearer valid-token
```

**请求示例**

```bash
curl -H "Authorization: Bearer valid-token" http://localhost:9002/api/v1/users
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": [
    { "id": "1", "name": "Alice" },
    { "id": "2", "name": "Bob" }
  ]
}
```

### 创建用户

```
POST /api/v1/users
```

**请求头**

```
Authorization: Bearer valid-token
Content-Type: application/json
```

**请求参数**

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 用户名，最少 2 个字符 |
| email | string | 是 | 邮箱，需符合邮箱格式 |

**请求示例**

```bash
curl -X POST \
  -H "Authorization: Bearer valid-token" \
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

**响应示例（验证失败）**

```json
{
  "code": 400,
  "message": "数据验证失败"
}
```

### 获取用户详情

```
GET /api/v1/users/:id
```

**请求头**

```
Authorization: Bearer valid-token
```

**路径参数**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | string | 用户 ID |

**请求示例**

```bash
curl -H "Authorization: Bearer valid-token" http://localhost:9002/api/v1/users/1
```

**响应示例**

```json
{
  "code": 0,
  "message": "操作成功",
  "data": { "id": "1" }
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

### 获取 Token

测试环境使用固定 Token: `valid-token`

在生产环境中，需要实现完整的 JWT 认证流程：

1. 用户登录获取 Token
2. 在请求头中携带 `Authorization: Bearer <token>`
3. 后端验证 Token 有效性

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
