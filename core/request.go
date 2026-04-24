package core

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

var validate = validator.New()

type Request struct {
	ctx fiber.Ctx
}

func NewRequest(c fiber.Ctx) *Request {
	return &Request{ctx: c}
}

// 获取单个字段（query / body / param 自动合并）
func (r *Request) Input(key string, defaultVal ...string) string {
	if val := r.ctx.Params(key); val != "" {
		return val
	}
	if val := r.ctx.Query(key); val != "" {
		return val
	}
	if val := r.ctx.FormValue(key); val != "" {
		return val
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// 获取所有输入（body JSON 绑定到 map）
func (r *Request) All() map[string]interface{} {
	data := make(map[string]interface{})
	_ = r.ctx.Bind().Body(&data)
	return data
}

// 绑定并验证结构体（tag: validate:"required,email"）
func (r *Request) Validate(out interface{}) error {
	if err := r.ctx.Bind().Body(out); err != nil {
		return err
	}
	return validate.Struct(out)
}

// 获取 Header
func (r *Request) Header(key string) string {
	return r.ctx.Get(key)
}

// 获取认证用户（从 Locals 注入，由 Auth 中间件设置）
func (r *Request) User() interface{} {
	return r.ctx.Locals("user")
}

// Bearer Token
func (r *Request) BearerToken() string {
	auth := r.ctx.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		return auth[7:]
	}
	return ""
}

// IP
func (r *Request) IP() string {
	return r.ctx.IP()
}
