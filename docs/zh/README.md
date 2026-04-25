# PureCore 框架

## 项目介绍

PureCore 是一个基于 Go 语言的全栈 Web 开发框架，将 GoFiber v3 封装成类似 Laravel 的开发风格，提供了路由分组、中间件管道、请求验证、统一响应格式、多语言支持等开箱即用的功能。前端采用 Vue 3 + Vite + Tailwind CSS + DaisyUI 技术栈。

- GitHub: [https://github.com/zhuchunshu/PureCore](https://github.com/zhuchunshu/PureCore)

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端语言 | Go |
| 后端框架 | GoFiber v3 |
| 数据库 | PostgreSQL |
| ORM | GORM |
| 验证 | go-playground/validator |
| CLI | Cobra（Artisan 风格） |
| 前端框架 | Vue 3 |
| 构建工具 | Vite |
| CSS 框架 | Tailwind CSS + DaisyUI |
| 包管理器 | Bun |

## 项目结构

```
/purecore
├── cmd/                   # CLI 命令
│   ├── root.go            # 根命令
│   ├── serve.go           # HTTP 服务器命令
│   └── migrate.go         # 数据库迁移命令
├── core/                  # 核心框架
│   ├── router.go          # Laravel 风格路由(分组/前缀/中间件)
│   ├── request.go         # 请求处理(Input/Validate/User)
│   ├── response.go        # 统一响应(Success/Error/Paginate)
│   ├── middleware.go       # HandlerFunc 类型 和 H() 桥接函数
│   ├── lang.go            # 多语言管理器
│   ├── database.go        # 数据库连接（GORM）
│   └── model.go           # 基础模型结构体
├── app/
│   ├── Http/
│   │   ├── Controllers/   # 应用控制器
│   │   │   ├── UserController.go
│   │   │   └── SystemController.go
│   │   └── Middleware/     # 中间件
│   │       ├── Auth.go    # Token 鉴权
│   │       ├── Cors.go    # 跨域处理
│   │       └── Lang.go    # 语言检测
│   └── Models/            # 数据库模型（GORM）
│       └── User.go
├── routes/                # 路由注册
│   └── api.go
├── lang/                  # 多语言翻译文件(前后端共享)
│   ├── zh.json            # 中文翻译
│   └── en.json            # 英文翻译
├── web/                   # 前端项目（Vue 3 + SSR，基于 Bun）
│   ├── src/
│   │   ├── i18n.js        # 前端多语言模块
│   │   ├── entry-client.js # 客户端入口（hydration）
│   │   ├── entry-server.js # 服务端入口（SSR）
│   │   ├── App.vue
│   │   └── main.js
│   ├── server.js          # SSR 服务器
│   ├── public/
│   │   └── lang/          # -> ../../lang (软链接)
│   ├── vite.config.js
│   └── package.json
├── .env                   # 环境配置(前后端共享)
├── main.go                # 后端入口
├── go.mod
└── go.sum
```

## 快速开始

### 环境要求

- Go 1.21+
- PostgreSQL
- Bun(或 Node.js + npm)
- Git

### 1. 克隆项目

```bash
git clone https://github.com/zhuchunshu/PureCore.git
cd PureCore
```

### 2. 配置环境

```bash
cp .env.example .env
# 编辑 .env 文件，修改数据库连接和端口配置
```

### 3. 编译并启动后端

```bash
# 安装依赖
go mod tidy

# 创建前端软链接(多语言文件共享)
ln -s ../../lang web/public/lang

# 编译二进制文件
go build -o purecore .

# 运行数据库迁移
./purecore migrate

# 启动服务(默认端口 9002)
./purecore serve
```

### 4. 启动前端

```bash
cd web
bun install
bun run dev
# 前端运行在 http://localhost:9001
# API 请求自动代理到 http://localhost:9002
```

## CLI 命令

| 命令 | 描述 |
|------|------|
| `./purecore serve` | 启动 HTTP 服务器 |
| `./purecore migrate` | 运行数据库迁移（根据模型自动创建数据表） |
| `./purecore --help` | 查看可用命令 |

## 核心特性

### 数据库与模型

PureCore 使用 **GORM** 作为 ORM，提供类似 Eloquent 的使用体验。基础 `core.Model` 结构体提供了 ID、时间戳和软删除支持：

```go
// app/Models/User.go
type User struct {
    core.Model
    Name  string `gorm:"type:varchar(100);not null" json:"name" validate:"required,min=2"`
    Email string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email" validate:"required,email"`
}
```

**模型约定：**
- 模型放在 `app/Models/` 目录下
- 嵌入 `core.Model` 获得 ID、CreatedAt、UpdatedAt、DeletedAt 字段
- 添加或修改模型后运行 `./purecore migrate` 即可创建对应的数据表
- 通过 `core.DB()` 访问数据库连接（单例模式）

### Laravel 风格路由

```go
// 路由分组、前缀和中间件
r.Prefix("/api/v1").Middleware(middleware.Auth()).Group(func(r *core.Router) {
    r.Get("/users", core.H(userCtrl.Index))
    r.Post("/users", core.H(userCtrl.Store))
    r.Get("/users/:id", core.H(userCtrl.Show))
})
```

### 请求验证

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

func (uc *UserController) Store(req *core.Request, res *core.Response) error {
    var body CreateUserRequest
    if err := req.Validate(&body); err != nil {
        return res.Error(err.Error())
    }
    // 处理业务逻辑...
}
```

### 统一响应

```go
res.Success(data)           // {"code":0,"message":"操作成功","data":{...}}
res.Error("错误信息")        // {"code":400,"message":"错误信息"}
res.NotFound("用户不存在")   // {"code":404,"message":"用户不存在"}
res.Unauthorized()          // {"code":401,"message":"未授权"}
res.Paginate(data, total, page, perPage) // 分页响应
```

### 多语言支持

前后端共用 `lang/` 目录下的 JSON 翻译文件。默认语言为**中文**。

- **后端**: 通过 `Accept-Language` 请求头自动检测。响应消息自动使用对应语言。
- **前端**: 通过浏览器语言自动检测。使用 `t('common.success')` 获取翻译。

```javascript
// 前端使用
import { t, setLocale } from './i18n'
console.log(t('common.success'))  // "操作成功"
setLocale('en')
console.log(t('common.success'))  // "Operation successful"
```

添加新语言只需在 `lang/` 目录下创建新的 JSON 文件即可。

## 配置说明

项目使用 `.env` 文件进行配置，前后端共享同一份配置。

| 变量 | 说明 | 默认值 |
|------|------|--------|
| FRONTEND_PORT | 前端开发服务器端口 | 9001 |
| BACKEND_PORT | 后端 API 服务器端口 | 9002 |
| DB_HOST | 数据库主机地址 | localhost |
| DB_PORT | 数据库端口 | 5432 |
| DB_USER | 数据库用户名 | postgres |
| DB_PASSWORD | 数据库密码 | postgres |
| DB_NAME | 数据库名称 | purecore |
| APP_ENV | 运行环境 | local |
| APP_DEBUG | 调试模式 | true |

## API 接口

详细接口文档请参阅 [API 文档](./API.md)。

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/ping | 健康检查 |
| GET | /api/v1/system/info | 项目信息 |

### 需认证接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /api/v1/users | 用户列表 |
| POST | /api/v1/users | 创建用户 |
| GET | /api/v1/users/:id | 用户详情 |

认证方式: `Authorization: Bearer <token>`

## 开发指南

详细开发文档请参阅 [开发文档](./DEVELOPMENT.md)。
