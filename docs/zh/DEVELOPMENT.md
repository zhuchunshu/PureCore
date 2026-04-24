# PureCore 开发文档

## 开发环境搭建

### 环境要求

- Go 1.21+
- PostgreSQL 15+
- Bun 1.0+ (或 Node.js 20 + npm)
- Git

### 项目初始化

```bash
# 克隆项目
git clone https://github.com/zhuchunshu/PureCore.git
cd PureCore

# 配置环境变量
cp .env.example .env

# 安装后端依赖
go mod tidy

# 创建多语言文件软链接 (前后端共享)
ln -s ../../lang web/public/lang

# 安装前端依赖
cd web && bun install
```

### 启动开发服务器

```bash
# 终端 1 - 启动后端 (默认端口 9002)
go run main.go

# 终端 2 - 启动前端 (默认端口 9001)
cd web && bun run dev
```

## 项目架构

### 核心框架 (core/)

#### router.go - 路由管理

提供 Laravel 风格的路由定义方式，支持链式调用：

```go
// 创建路由组
r := core.NewRouter(app)

// 公开路由
r.Prefix("/api/v1").Group(func(r *core.Router) {
    r.Get("/ping", handler)
})

// 需要认证的路由组
r.Prefix("/api/v1").Middleware(authMiddleware).Group(func(r *core.Router) {
    r.Get("/users", userHandler)
    r.Post("/users", userHandler)
})
```

支持的方法: `Get`, `Post`, `Put`, `Delete`, `Patch`

#### request.go - 请求处理

```go
// 获取单个输入字段 (自动合并 param/query/body)
name := req.Input("name")
name := req.Input("name", "默认值")

// 获取所有输入
allInput := req.All()

// 结构体绑定与验证
var body CreateUserRequest
if err := req.Validate(&body); err != nil {
    // 验证失败处理
}

// 获取当前认证用户
user := req.User()

// 获取 Bearer Token
token := req.BearerToken()

// 获取请求头
header := req.Header("Content-Type")

// 获取客户端 IP
ip := req.IP()
```

#### response.go - 响应处理

```go
// 成功响应
res.Success(data)                    // 200, code=0

// 错误响应
res.Error("错误信息")                // 400
res.Error("错误信息", 422)           // 自定义状态码

// 未授权
res.Unauthorized()                   // 401

// 资源不存在
res.NotFound()                       // 404
res.NotFound("用户不存在")           // 自定义消息

// 分页响应
res.Paginate(data, total, page, perPage)

// 自定义 JSON
res.JSON(status, code, message, data)
```

响应消息自动根据 `Accept-Language` 请求头使用对应语言翻译。

#### middleware.go - 中间件桥接

```go
// Controller 方法签名
type HandlerFunc func(req *Request, res *Response) error

// H() 将 HandlerFunc 转换为 fiber.Handler
r.Get("/path", core.H(func(req *core.Request, res *core.Response) error {
    return res.Success("hello")
}))

// 或绑定到 Controller 方法
r.Get("/users", core.H(userCtrl.Index))
```

#### lang.go - 多语言管理

```go
// 初始化语言文件 (在 main.go 中调用)
core.InitLang("lang")

// 获取翻译
msg := core.GetLang().Trans("common.success")

// 设置当前语言
core.GetLang().SetLocale("en")

// 获取当前语言
locale := core.GetLang().GetLocale()
```

### 应用层 (app/Http/)

#### Controllers

控制器放在 `app/Http/Controllers/` 目录下，方法签名为 `func(req *core.Request, res *core.Response) error`：

```go
package controllers

import "purecore/core"

type UserController struct{}

func (uc *UserController) Index(req *core.Request, res *core.Response) error {
    // 返回用户列表
    return res.Success(users)
}

func (uc *UserController) Store(req *core.Request, res *core.Response) error {
    // 验证请求数据
    var body CreateUserRequest
    if err := req.Validate(&body); err != nil {
        return res.Error(err.Error())
    }
    // 创建用户
    return res.Success(newUser)
}

func (uc *UserController) Show(req *core.Request, res *core.Response) error {
    id := req.Input("id")
    // 查找用户
    return res.Success(user)
}
```

#### Middleware

中间件放在 `app/Http/Middleware/` 目录下：

```go
package middleware

import "github.com/gofiber/fiber/v3"

func MyMiddleware() fiber.Handler {
    return func(c fiber.Ctx) error {
        // 前置处理
        err := c.Next()
        // 后置处理
        return err
    }
}
```

在路由中使用：
```go
r.Prefix("/api/v1").Middleware(middleware.MyMiddleware()).Group(func(r *core.Router) {
    // 受保护的路由
})
```

### 路由注册 (routes/)

在 `routes/api.go` 中注册路由：

```go
package routes

func RegisterAPI(r *core.Router) {
    // 公开路由
    r.Prefix("/api/v1").Group(func(r *core.Router) {
        r.Get("/ping", ...)
    })

    // 认证路由
    r.Prefix("/api/v1").Middleware(middleware.Auth()).Group(func(r *core.Router) {
        r.Get("/users", ...)
        r.Post("/users", ...)
    })
}
```

在 `main.go` 中调用：
```go
router := core.NewRouter(app)
routes.RegisterAPI(router)
```

## 多语言支持

### 翻译文件结构

翻译文件位于 `lang/` 目录，采用 JSON 格式，使用嵌套结构：

```json
{
  "common": {
    "success": "操作成功",
    "error": "操作失败",
    "not_found": "资源不存在",
    "unauthorized": "未授权"
  },
  "auth": {
    "login_success": "登录成功",
    "token_invalid": "令牌无效"
  }
}
```

### 访问翻译

按键使用点号分隔：`"common.success"`, `"auth.login_success"`

### 后端使用

```go
// 自动检测 Accept-Language 请求头
// 默认语言为中文 (zh)

// 手动设置语言
core.GetLang().SetLocale("en")

// 获取翻译
msg := core.GetLang().Trans("common.success")
```

### 前端使用

```javascript
import { t, setLocale, initI18n } from './i18n'

// 初始化 (自动检测浏览器语言)
await initI18n()

// 获取翻译
t('common.success')  // 中文: "操作成功", 英文: "Operation successful"

// 切换语言
await setLocale('en')
```

### 添加新语言

1. 在 `lang/` 目录创建新的 JSON 文件 (如 `ja.json`)
2. 按照现有文件格式填写翻译
3. 前后端自动加载，无需额外配置

## 配置管理

### .env 文件

```env
# 前端
FRONTEND_PORT=9001

# 后端
BACKEND_PORT=9002

# 数据库
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=purecore
DB_SSLMODE=disable

# 应用
APP_ENV=local
APP_DEBUG=true
```

### Vite 配置

前端 `web/vite.config.js` 读取 `.env` 配置：

```javascript
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  return {
    server: {
      port: parseInt(env.FRONTEND_PORT) || 9001,
      proxy: {
        '/api': {
          target: `http://localhost:${env.BACKEND_PORT || 9002}`,
          changeOrigin: true,
        },
      },
    },
  }
})
```

## 数据库配置

项目使用 PostgreSQL 数据库。在 `.env` 文件中配置数据库连接信息。

```go
// 数据库连接示例 (使用 GORM)
import (
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
    os.Getenv("DB_HOST"),
    os.Getenv("DB_USER"),
    os.Getenv("DB_PASSWORD"),
    os.Getenv("DB_NAME"),
    os.Getenv("DB_PORT"),
    os.Getenv("DB_SSLMODE"),
)

db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

## 部署

### 构建Go 后端

```bash
go build -o purecore-server .
./purecore-server
```

### 构建前端

```bash
cd web
bun run build
# 静态文件输出到 web/dist/
```

## 常见问题

### 1. 端口冲突

修改 `.env` 文件中的 `FRONTEND_PORT` 和 `BACKEND_PORT` 即可。

### 2. 找不到 lang/ 文件

确保已创建软链接：
```bash
ln -s ../../lang web/public/lang
```

### 3. Go 模块问题

```bash
go clean -modcache
go mod tidy
