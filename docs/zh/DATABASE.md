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

运行 migration 命令，根据模型创建或更新数据表：

```bash
./purecore migrate
```

**工作原理：**
- `cmd/migrate.go` 包含要迁移的模型列表
- GORM 的 `AutoMigrate` 创建缺失的表和列
- 不会破坏现有数据——迁移是无损的
- 新模型必须在 `cmd/migrate.go` 的 `models` 切片中注册

**添加新模型到迁移：**

```go
// 在 cmd/migrate.go 中
func migrateRun(cmd *cobra.Command, args []string) {
    db := core.DB()

    models := []interface{}{
        &models.User{},
        &models.Post{},    // ← 在这里添加新模型
    }

    for _, model := range models {
        name := reflect.TypeOf(model).Elem().Name()
        if err := db.AutoMigrate(model); err != nil {
            log.Fatalf("Failed to migrate %s: %v", name, err)
        }
        log.Printf("  ✓ %s table migrated", name)
    }
}
```

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
