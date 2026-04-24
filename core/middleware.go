package core

import "github.com/gofiber/fiber/v3"

// 中间件类型别名，让外层更语义化
type MiddlewareFunc = fiber.Handler

// 封装 Controller Handler，自动注入 Request/Response
type HandlerFunc func(req *Request, res *Response) error

// 将 HandlerFunc 转换为 fiber.Handler
func H(fn HandlerFunc) fiber.Handler {
	return func(c fiber.Ctx) error {
		req := NewRequest(c)
		res := NewResponse(c)
		return fn(req, res)
	}
}
