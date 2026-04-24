package controllers

import "purecore/core"

type UserController struct{}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

func (uc *UserController) Index(req *core.Request, res *core.Response) error {
	users := []map[string]string{
		{"id": "1", "name": "Alice"},
		{"id": "2", "name": "Bob"},
	}
	return res.Success(users)
}

func (uc *UserController) Store(req *core.Request, res *core.Response) error {
	var body CreateUserRequest
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error())
	}
	return res.Success(map[string]string{
		"name":  body.Name,
		"email": body.Email,
	})
}

func (uc *UserController) Show(req *core.Request, res *core.Response) error {
	id := req.Input("id")
	if id == "" {
		return res.NotFound("用户不存在")
	}
	return res.Success(map[string]string{"id": id})
}
