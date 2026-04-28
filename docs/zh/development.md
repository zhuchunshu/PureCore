# 开发指南

## 环境要求

- Go 1.21+
- PostgreSQL（用于数据持久化）

## 安装与运行

### 安装依赖


### 配置环境变量

复制 `.env.example` 为 `.env`（如存在）或直接设置环境变量：


### 运行开发服务器


服务器默认运行在 `http://localhost:3010`

## 添加新功能

### 1. 创建控制器

在 `app/Http/Controllers/` 目录下创建新文件，例如 `ProductController.go`：


### 2. 添加路由

在 `routes/api.go` 或 `routes/web.go` 中注册路由：


`ApiResource` 会自动注册以下路由：

| 方法 | 路径 | 控制器方法 |
|------|------|------------|
| GET | /products | Index |
| GET | /products/{id} | Show |
| POST | /products | Store |
| PUT | /products/{id} | Update |
| DELETE | /products/{id} | Destroy |

### 3. 创建中间件

在 `app/Http/Middleware/` 下创建中间件：


在路由中应用中间件：


需要在某处注册中间件名称（例如在 `RouteServiceProvider` 中）：


### 4. 创建服务提供者

在 `app/Providers/` 下创建新提供者：


在 `bootstrap/app.go` 中注册：


### 5. 使用 Facade


## 请求处理

### 获取输入参数


### 响应格式


## 数据库操作

框架已集成 **GORM**（PostgreSQL 驱动），封装了类似 Laravel 的 **DB 门面**、**Eloquent 模型** 和 **数据迁移** 系统。

### 配置数据库

编辑项目根目录的 `.env` 文件（首次启动时自动从 `.env.example` 复制）：

```env
DB_CONNECTION=pgsql
DB_HOST=127.0.0.1
DB_PORT=5432
DB_DATABASE=purecore
DB_USERNAME=purecore
DB_PASSWORD=
```

应用启动时自动调用 `core.InitDB()` 和 `core.RunMigrations()`。

### DB 门面

全局变量 `core.DB` 提供 GORM 实例，可在任意位置直接使用：

```go
import "purecore/core"

// 查询多条
var users []User
core.DB.Where("age > ?", 18).Find(&users)

// 查询单条
var user User
core.DB.First(&user, 1)

// 创建
core.DB.Create(&User{Name: "Alice", Email: "alice@example.com"})

// 更新
core.DB.Model(&user).Update("Name", "Bob")

// 删除（软删除）
core.DB.Delete(&user)
```

也可通过 `core.GetDB()` 获取实例。

### 模型定义

模型应嵌入 `core.Model`，获得主键、时间戳和软删除支持：

```go
package models

import "purecore/core"

type User struct {
    core.Model
    Name  string `gorm:"size:100;not null" json:"name"`
    Email string `gorm:"uniqueIndex;size:255;not null" json:"email"`
}
```

`core.Model` 包含以下字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| ID | uint | 自增主键 |
| CreatedAt | time.Time | 创建时间 |
| UpdatedAt | time.Time | 更新时间 |
| DeletedAt | gorm.DeletedAt | 软删除时间（索引） |

### 数据迁移

类似 Laravel Migration，支持版本控制和回滚。

#### 定义迁移

```go
package migrations

import "purecore/core"

type CreateUsersTable struct{}

func (m *CreateUsersTable) Up() error {
    type User struct {
        core.Model
        Name  string `gorm:"size:100"`
        Email string `gorm:"uniqueIndex"`
    }
    return core.DB.AutoMigrate(&User{})
}

func (m *CreateUsersTable) Down() error {
    return core.DB.Migrator().DropTable("users")
}
```

#### 注册迁移

```go
func init() {
    core.RegisterMigration("create_users_table", &CreateUsersTable{})
}
```

> 迁移按注册名称的字母顺序执行。推荐使用时间戳前缀（如 `20240101_create_users_table`）确保顺序。

#### 迁移命令

使用 CLI 命令管理迁移：

```bash
# 运行迁移（启动服务时自动执行）
go run . migrate

# 回滚最近一批
go run . migrate:rollback

# 回滚全部并重新执行
go run . migrate:refresh

# 启动 Web 服务（自动运行迁移）
go run . serve
```

迁移记录存储在 `migrations` 表中，自动跟踪批次和执行状态。

### 完整示例

在控制器中使用数据库：

```go
package controllers

import (
    "purecore/core"
    "purecore/app/models"
)

type UserController struct{}

func (uc *UserController) Index(req *core.Request, res *core.Response) error {
    var users []models.User
    core.DB.Find(&users)
    return res.Success(users)
}

func (uc *UserController) Store(req *core.Request, res *core.Response) error {
    var user models.User
    if err := req.Validate(&user); err != nil {
        return res.Error(err.Error())
    }
    core.DB.Create(&user)
    return res.Success(user)
}
```

## 测试


## 代码规范

- 包名使用小写单词
- 文件名使用 PascalCase（如 `UserController.go`）
- 接口名以 `er` 结尾（如 `ServiceProvider`）
- 错误处理使用 Go 标准方式（返回 error）
- 避免全局变量，优先依赖注入
