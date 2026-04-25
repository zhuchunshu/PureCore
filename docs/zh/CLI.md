# CLI 命令

PureCore 提供由 [Cobra](https://github.com/spf13/cobra) 驱动的 Artisan 风格命令行界面。

## 可用命令

| 命令 | 描述 |
|------|------|
| `./purecore serve` | 启动 HTTP 服务器 |
| `./purecore make:model` | 创建新的模型文件 |
| `./purecore make:controller` | 创建新的控制器文件 |
| `./purecore make:migration` | 创建新的迁移文件 |
| `./purecore migrate` | 运行数据库迁移 |
| `./purecore --help` | 查看所有可用命令 |
| `./purecore completion` | 生成 Shell 自动补全脚本 |

## serve

启动 PureCore HTTP 服务器。

```bash
./purecore serve
```

服务器监听由 `BACKEND_PORT` 环境变量指定的端口（默认：`9002`）。

**启动过程：**
1. 加载 `.env` 环境变量
2. 初始化语言管理器（`lang/` 目录）
3. 建立数据库连接（GORM + PostgreSQL）
4. 注册中间件（CORS、语言检测）
5. 注册所有 API 路由
6. 开始监听 HTTP 请求

## migrate

运行所有注册但尚未执行的数据库迁移。

```bash
./purecore migrate
```

**工作原理：**
- 使用 `.env` 中的凭据连接数据库
- 如果 `migrations` 表不存在，则创建该表以跟踪执行历史
- 遍历 `database/migrations/` 中通过 `init()` 注册的所有迁移
- 对每个待处理的迁移执行 GORM 的 `AutoMigrate`
- 在数据库中记录每个迁移，防止重复执行

迁移通过 Go 的 `init()` 机制自动包含在二进制文件中——无需扫描文件系统。`cmd/serve.go` 文件导入了 migrations 包，因此所有注册的迁移在编译时已嵌入，并在服务器启动时自动运行。

**使用 make 命令添加新模型：**

```bash
./purecore make:model Post
./purecore make:migration Post
```

然后重新编译并运行：

```bash
go build -o purecore .
./purecore migrate
```

## make:model

创建一个新的 GORM 模型文件到 `app/Models/`。

```bash
./purecore make:model Post
```

这会生成 `app/Models/Post.go`，包含：
- 包声明和 `purecore/core` 导入
- 嵌入 `core.Model` 的结构体
- 带有 GORM 和验证标签的 `Name` 字段

**创建模型后：**
1. 将模型添加到 `cmd/migrate.go` 的迁移列表中
2. 运行 `./purecore migrate` 创建数据库表

## make:controller

创建一个新的控制器文件到 `app/Http/Controllers/`，包含完整的 CRUD 脚手架代码。

```bash
./purecore make:controller Post
```

这会生成 `app/Http/Controllers/PostController.go`，包含：
- `Index` — 列出所有记录
- `Store` — 创建新记录（含验证）
- `Show` — 根据 ID 获取单条记录

每个方法使用 `app/Models/` 中对应的模型，并通过 `core.DB()` 访问数据库。

## make:migration

创建一个新的迁移文件到 `database/migrations/`。

```bash
./purecore make:migration Post
```

这会生成一个迁移文件，包含：
- `init()` 注册 — 调用 `core.RegisterMigration()` 自动注册迁移
- `up()` 函数 — 使用 GORM AutoMigrate 创建表，嵌入 `core.Model`

迁移在编译时自动注册——无需手动添加到任何列表。只需重新编译并运行 `./purecore migrate` 即可。

## 添加新命令

1. 在 `cmd/` 中创建新文件（例如 `cmd/mycommand.go`）
2. 在 `init()` 中用 `rootCmd.AddCommand(mycmd)` 注册
3. 重新编译：`go build -o purecore .`

```go
// cmd/mycommand.go
package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var myCmd = &cobra.Command{
    Use:   "mycommand",
    Short: "命令描述",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Hello from my command!")
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
