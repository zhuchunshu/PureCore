package core

import (
	"github.com/gofiber/fiber/v3"
)

type Response struct {
	ctx fiber.Ctx
}

func NewResponse(c fiber.Ctx) *Response {
	return &Response{ctx: c}
}

type JsonResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PaginateResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
}

func (r *Response) JSON(status int, code int, message string, data interface{}) error {
	return r.ctx.Status(status).JSON(JsonResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func (r *Response) Success(data interface{}) error {
	return r.JSON(200, 0, GetLang().Trans("common.success"), data)
}

func (r *Response) Error(message string, code ...int) error {
	c := 400
	if len(code) > 0 {
		c = code[0]
	}
	return r.ctx.Status(c).JSON(JsonResponse{
		Code:    c,
		Message: message,
	})
}

func (r *Response) Unauthorized() error {
	return r.Error(GetLang().Trans("common.unauthorized"), 401)
}

func (r *Response) NotFound(message ...string) error {
	msg := GetLang().Trans("common.not_found")
	if len(message) > 0 {
		msg = message[0]
	}
	return r.Error(msg, 404)
}

func (r *Response) Paginate(data interface{}, total int64, page int, perPage int) error {
	return r.ctx.Status(200).JSON(PaginateResponse{
		Code:    0,
		Message: "success",
		Data:    data,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	})
}
