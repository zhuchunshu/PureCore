package middleware

import (
	"purecore/core"

	"github.com/gofiber/fiber/v3"
)

func Auth() fiber.Handler {
	return func(c fiber.Ctx) error {
		req := core.NewRequest(c)
		res := core.NewResponse(c)
		token := req.BearerToken()
		if token == "" || token != "valid-token" { // 替换为真实 JWT 验证
			return res.Unauthorized()
		}
		// 注入用户到 Locals，后续 req.User() 可取到
		c.Locals("user", map[string]string{"id": "1", "name": "Alice"})
		return c.Next()
	}
}
