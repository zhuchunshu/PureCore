# 数据库与模型

PureCore 使用 **[GORM](https://gorm.io/)** 作为 ORM，为数据库操作提供类似 Eloquent 的体验。

## 配置

数据库设置通过 `.env` 文件中的环境变量进行配置：

| 变量 | 描述 | 默认值 |
|------|------|--------|
| `DB_HOST` | 数据库主机地址 | `localhost` |
| `DB_PORT` | 数据库端口 | `5432` |
| `DB_USER` | 数据库用户名 | `postgres` |
| `DB_PASSWORD` | 数据库密码 | `postgres` |
| `DB_NAME` | 数据库名称 | `purecore` |
| `DB_SSLMODE` | SSL 模式 | `disable` |

## 基础模型

所有数据库模型都应嵌入 `core.Model` 以获得标准字段：

```go
// core/model.go
type Model struct {
    ID        uint           `gorm:"primarykey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

## 创建模型

模型文件放在 `app/Models/` 目录下。以下是一个例子：

```go
// app/Models/User.go
package models

import "purecore/core"

type User struct {
    core.Model
    Name  string `gorm:"type:varchar(100);not null" json:"name" validate:"required,min=2"`
    Email string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" validate:"required,email"`
}
```

**GORM 结构体标签参考：**

| 标签 | 用途 |
|------|------|
| `primarykey` | 指定主键 |
| `type:varchar(n)` | 设置列类型和长度 |
| `not null` | 禁止 NULL 值 |
| `uniqueIndex` | 创建唯一索引 |
| `default:value` | 设置默认值 |
| `autoIncrement` | 自增整数 |

## 访问数据库

在应用程序的任何地方使用 `core.DB()` 获取单例数据库连接：

```go
import (
    models "purecore/app/Models"
    "purecore/core"
)

// 创建
user := models.User{Name: "张三", Email: "zhangsan@example.com"}
core.DB().Create(&user)

// 读取（单条）
var user models.User
core.DB().First(&user, 1)        // 按主键
core.DB().Where("email = ?", "zhangsan@example.com").First(&user)

// 读取（列表）
var users []models.User
core.DB().Find(&users)
core.DB().Where("name LIKE ?", "%张%").Find(&users)

// 更新
core.DB().Model(&user).Update("name", "张三（已更新）")
core.DB().Model(&user).Updates(models.User{Name: "新名字", Email: "new@example.com"})

// 删除（软删除）
core.DB().Delete(&user)
```

## 数据库迁移

PureCore 使用类似 Laravel 的迁移系统，每个迁移通过 `init()` 注册并在数据库中跟踪。迁移在服务器启动时自动运行，也可以手动触发：

```bash
./purecore migrate
```

**工作原理：**
- `database/migrations/` 中的每个迁移文件通过 `init()` 使用 `core.RegisterMigration(name, upFunc)` 注册自身
- `core.RunMigrations()`（位于 `core/migrator.go`）读取所有已注册的迁移并运行尚未执行的迁移
- 数据库中的 `migrations` 表跟踪已运行的迁移及其批次号
- 迁移编译进二进制文件——无需文件系统扫描
- `cmd/serve.go` 文件导入迁移包（`_ "purecore/database/migrations"`），因此所有已注册的迁移都可用

**示例迁移文件：**

```go
// database/migrations/2026_04_27_175200_create_web_options_table.go
package migrations

import (
    "purecore/core"
    "gorm.io/gorm"
)

func init() {
    core.RegisterMigration("2026_04_27_175200_create_web_options_table", up2026_04_27_175200)
}

func up2026_04_27_175200(db *gorm.DB) error {
    type WebOption struct {
        core.Model
        Name  string `gorm:"type:varchar(100);uniqueIndex;not null"`
        Value string `gorm:"type:text;not null;default:''"`
    }
    return db.AutoMigrate(&WebOption{})
}
```

**使用 CLI 创建迁移：**

```bash
./purecore make:migration create_posts_table
```

这会生成一个迁移文件，包含正确的 `init()` 注册和 `up()` 函数。迁移会自动注册——只需重新构建并运行 `./purecore migrate`。

**迁移命名规范：**
- 使用时间戳前缀以确保顺序：`YYYY_MM_DD_HHMMSS_description.go`
- 示例：`2026_04_30_120000_create_posts_table.go`

**迁移跟踪：**
- `migrations` 表存储迁移名称、批次号和时间戳
- 每次 `migrate` 运行为所有执行的迁移创建新的批次号
- 通过在跟踪表中查找，跳过之前已运行的迁移

## 查询构建器

GORM 提供流畅的查询构建器：

```go
// Where 条件
db.Where("age > ?", 18).Find(&users)
db.Where("name = ? AND email = ?", name, email).Find(&users)

// 排序
db.Order("created_at desc").Find(&users)

// 限制和偏移
db.Limit(10).Offset(20).Find(&users)

// 计数
var count int64
db.Model(&models.User{}).Where("active = ?", true).Count(&count)

// 事务
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&user1).Error; err != nil {
        return err
    }
    if err := tx.Create(&user2).Error; err != nil {
        return err
    }
    return nil
})
```

## 调试

在 `.env` 文件中设置 `APP_DEBUG=true` 以启用 SQL 查询日志记录。这会将所有 SQL 查询输出到控制台。
