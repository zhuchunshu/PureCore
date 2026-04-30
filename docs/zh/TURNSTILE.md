# Turnstile 集成

PureCore 支持 [Cloudflare Turnstile](https://developers.cloudflare.com/turnstile/) —— 一种注重隐私的 CAPTCHA 替代方案 —— 用于保护面向公众的认证端点免受机器人和自动化滥用。

## 概述

Turnstile 验证可以为以下场景按上下文启用：

| 上下文键 | 受保护的端点 |
|----------|-------------|
| `turnstile_public_login` | `/api/v1/auth/login`、`/api/v1/auth/register` |
| `turnstile_admin_login` | `/{admin_prefix}/auth/login` |
| `turnstile_admin_register` | `/{admin_prefix}/auth/register` |

每个上下文都可以通过管理员选项独立切换。

## 工作原理

### 流程

1. 当相应选项启用时，前端渲染 Turnstile 组件（使用 `TurnstileWidget.vue`）
2. 用户完成 Turnstile 验证
3. 提交表单时，Turnstile 令牌与其他表单数据一并发送（例如 `turnstile_token` 字段）
4. 后端调用 Cloudflare 的 siteverify 端点验证令牌
5. 如果验证成功，请求继续处理；否则，返回 422 错误

### 后端实现

Turnstile 逻辑集中位于 `core/turnstile.go`：

```go
// 验证 Turnstile 令牌
func VerifyTurnstile(token string) error

// 检查给定上下文是否启用了 Turnstile
func IsTurnstileEnabled(context string) bool

// 获取适当的站点密钥（开发模式返回测试密钥）
func GetTurnstileSiteKey() string
```

**控制器集成：**

```go
func (uc *UserAuthController) Login(req *core.Request, res *core.Response) error {
    var body UserLoginRequest
    if err := req.Validate(&body); err != nil {
        return res.Error("凭据无效", 422)
    }

    // 如果启用了公网登录 Turnstile，则进行验证
    if core.IsTurnstileEnabled("turnstile_public_login") {
        if err := core.VerifyTurnstile(body.TurnstileToken); err != nil {
            return res.Error("验证码验证失败: "+err.Error(), 422)
        }
    }
    // ... 继续认证流程
}
```

### 前端实现

`web/src/components/TurnstileWidget.vue` 组件：

- 动态加载 Cloudflare Turnstile 脚本
- 当站点密钥可用时渲染组件
- 验证成功完成后触发 `verified` 事件并附带令牌
- 支持重置组件以进行重新验证（例如表单提交失败后）

**在登录/注册表单中使用：**

```vue
<script setup>
import TurnstileWidget from '@/components/TurnstileWidget.vue'

const turnstileToken = ref('')

function handleSubmit() {
  // 将 turnstileToken.value 与表单数据一起发送
}
</script>

<template>
  <TurnstileWidget
    v-if="turnstileEnabled"
    @verified="token => turnstileToken = token"
  />
</template>
```

## 配置

### 管理选项

Turnstile 通过管理选项系统（`web_options` 表）进行配置。管理员可以通过管理员设置面板或 API 设置这些选项：

| 选项键 | 描述 | 示例值 |
|--------|------|--------|
| `turnstile_site_key` | Cloudflare Turnstile 站点密钥 | `1x00000000000000000000AA` |
| `turnstile_secret_key` | Cloudflare Turnstile 密钥 | `1x0000000000000000000000000000000AA` |
| `turnstile_public_login` | 为公网用户登录/注册启用 Turnstile | `"1"`（启用）或 `""`（禁用） |
| `turnstile_admin_login` | 为管理员登录启用 Turnstile | `"1"` 或 `""` |
| `turnstile_admin_register` | 为管理员注册启用 Turnstile | `"1"` 或 `""` |

**通过 API 设置选项：**

```bash
curl -X POST \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"turnstile_site_key":"1x00000000000000000000AA","turnstile_secret_key":"1x0000000000000000000000000000000AA","turnstile_public_login":"1"}' \
  http://localhost:9002/control-panel/options
```

### 检查 Turnstile 状态（公共 API）

前端可以通过公共选项端点检查 Turnstile 是否已启用并获取站点密钥：

```bash
curl http://localhost:9002/control-panel/options
```

如果已配置，响应中包含 `turnstile_site_key`。前端检查此密钥的存在以及相应的上下文切换，以决定是否渲染组件。

## 开发与测试

### 测试密钥

Cloudflare 提供**测试密钥**，始终通过验证，可在任何域名上使用，包括 `localhost`：

| 密钥类型 | 值 |
|----------|-----|
| 站点密钥 | `1x00000000000000000000AA` |
| 密钥 | `1x0000000000000000000000000000000AA` |

后端自动检测测试密钥（任何以 `1x` 开头的密钥），并在验证令牌时使用官方测试密钥。这意味着：

- 在开发期间将测试密钥设置为管理选项
- 组件将始终通过验证
- 使用测试密钥不会向 Cloudflare 的服务器发送请求

**开发环境配置示例：**

```
turnstile_site_key=1x00000000000000000000AA
turnstile_secret_key=1x0000000000000000000000000000000AA
turnstile_public_login=1
```

### 禁用 Turnstile

要禁用特定上下文的 Turnstile，可以：

1. 将上下文选项设置为空字符串或删除它
2. 删除站点密钥和密钥（禁用所有 Turnstile 功能）

## 安全说明

1. **必须进行服务端验证**：切勿信任客户端验证 Turnstile —— 始终在后端使用 `core.VerifyTurnstile()`
2. **密钥保护**：密钥存储在 `web_options` 数据库表中（仅认证管理员可访问）。切勿在客户端代码中暴露它。
3. **测试密钥检测**：后端仅在配置的密钥以 `1x` 开头时使用测试密钥。在生产环境中，始终使用真实的 Cloudflare 密钥。
4. **令牌一次性使用**：每个 Turnstile 令牌只能验证一次。表单每次提交尝试都必须生成新的令牌。
