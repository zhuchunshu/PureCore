# CLI 命令

PureCore 提供基于 [Cobra](https://github.com/spf13/cobra) 的 Artisan 风格 CLI。

## 可用命令

| 命令 | 描述 |
|------|------|
| `./purecore serve` | 启动 HTTP 服务器 |
| `./purecore migrate` | 运行数据库迁移 |
| `./purecore make:model` | 创建新的模型文件 |
| `./purecore make:controller` | 创建新的控制器文件 |
| `./purecore make:migration` | 创建新的迁移文件 |
| `./purecore make:module` | 创建新的服务提供者模块 |
| `./purecore make:middleware` | 创建新的中间件文件 |
| `./purecore --help` | 查看所有可用命令 |
| `./purecore completion` | 生成 Shell 自动补全脚本 |

## serve

启动 PureCore HTTP 服务器。

```bash
./purecore serve
```

服务器监听 `BACKEND_PORT` 环境变量指定的端口（默认：`9002`）。

**启动时发生的过程：**
1. 从 `.env` 加载环境变量
2. 初始化语言管理器（`lang/` 目录）
3. 建立数据库连接（GORM + PostgreSQL）
4. 加载并运行所有已注册的服务提供者（路由、中间件）
5. 运行所有待处理的数据库迁移
6. 开始监听 HTTP 请求

## migrate

运行所有已注册但尚未执行的数据库迁移。

```bash
./purecore migrate
```

**工作原理：**
- 使用 `.env` 中的凭据连接数据库
- 如果不存在，创建 `migrations` 表以跟踪执行历史
- 遍历 `database/migrations/` 中通过 `init()` 注册的所有迁移
- 对每个待处理的迁移运行 GORM 的 `AutoMigrate`
- 在数据库中记录每个迁移以防止重复执行

迁移通过 Go 的 `init()` 机制自动包含在二进制文件中——无需文件系统扫描。`cmd/serve.go` 文件导入迁移包，因此所有已注册的迁移都会在编译时包含，并在服务器启动时自动运行。

**使用 make 命令添加新模型：**

```bash
./purecore make:model Post
./purecore make:migration Post
```

然后重新构建并运行：

```bash
go build -o purecore .
./purecore migrate
```

## make:model

在 `app/Models/` 中创建新的 GORM 模型文件。

```bash
./purecore make:model Post
```

这会生成 `app/Models/Post.go`，包含：
- 包声明和 `purecore/core` 导入
- 嵌入 `core.Model` 的结构体
- 带有 GORM 和验证标签的 `Name` 字段

**创建模型后：**
1. 将模型添加到 `cmd/migrate.go` 中的迁移列表
2. 运行 `./purecore migrate` 创建数据库表

## make:controller

在 `app/Http/Controllers/` 中创建新的控制器文件，包含完整的 CRUD 脚手架。

```bash
./purecore make:controller Post
```

这会生成 `app/Http/Controllers/PostController.go`，包含：
- `Index` — 列出所有记录
- `Store` — 创建新记录（含验证）
- `Show` — 按 ID 获取单个记录

每个方法使用 `app/Models/` 中对应的模型，并通过 `core.DB()` 访问数据库。

## make:migration

在 `database/migrations/` 中创建新的迁移文件。

```bash
./purecore make:migration Post
```

这会生成一个迁移文件，包含：
- `init()` 注册 — 调用 `core.RegisterMigration()` 自动注册迁移
- `up()` 函数 — 使用 GORM AutoMigrate 创建表，嵌入 `core.Model`

迁移在编译二进制文件时自动注册——无需手动添加到任何列表。只需重新构建并运行 `./purecore migrate`。

## make:module

在 `app/Providers/` 中创建新的服务提供者模块。

```bash
./purecore make:module Post
```

这会生成 `app/Providers/PostServiceProvider.go`，实现 `core.ServiceProvider` 接口，包含：
- `Name()` — 返回提供者的唯一标识符（`"post"`）
- `Register(router)` — 添加路由和中间件的占位符
- `Boot()` — 注册后初始化钩子

**创建模块后：**

1. 将提供者添加到 `cmd/serve.go` 中的 `AddProviders()` 链：
```go
app.AddProviders(
    &providers.RouteServiceProvider{},
    &providers.PostServiceProvider{},  // ← 添加你的新提供者
)
```
2. 在 `Register()` 方法中实现你的路由
3. 在 `Boot()` 方法中添加任何初始化逻辑

**Register() 实现示例：**

```go
func (p *PostServiceProvider) Register(router *core.Router) error {
    router.Prefix("/api/v1").Group(func(r *core.Router) {
        r.Get("/posts", core.H(postCtrl.Index))
        r.Post("/posts", core.H(postCtrl.Store))
    })
    return nil
}
```

## make:middleware

在 `app/Http/Middleware/` 中创建新的中间件文件。

```bash
./purecore make:middleware RateLimit
```

这会生成 `app/Http/Middleware/RateLimit.go`，包含一个返回 `fiber.Handler` 的函数：

```go
package middleware

import "github.com/gofiber/fiber/v3"

func RateLimit() fiber.Handler {
    return func(c fiber.Ctx) error {
        // 在此处添加你的中间件逻辑
        return c.Next()
    }
}
```

**在路由中使用中间件：**

```go
// 直接使用
router.Middleware(middleware.RateLimit()).Group(func(r *core.Router) {
    r.Get("/protected", handler)
})

// 注册为命名中间件（在服务提供者中）：
router.RegisterNamedMiddlewares(map[string]core.MiddlewareFunc{
    "rate_limit": middleware.RateLimit(),
})

// 然后通过名称使用
router.MiddlewareByName("rate_limit").Group(func(r *core.Router) {
    r.Get("/protected", handler)
})
```

**对于全局中间件**，将其添加到 `cmd/serve.go` 中的 `RunWithMiddleware()`。

## 添加新命令

1. 在 `cmd/` 中创建新文件（例如 `cmd/mycommand.go`）
2. 在 `init()` 中使用 `rootCmd.AddCommand(mycmd)` 注册
3. 重新构建：`go build -o purecore .`

```go
// cmd/mycommand.go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "我的命令描述",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Hello from my command!")
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
