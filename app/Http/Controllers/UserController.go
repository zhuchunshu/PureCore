package controllers

import (
	models "purecore/app/Models"
	"purecore/core"
)

type UserController struct{}

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

func (uc *UserController) Index(req *core.Request, res *core.Response) error {
	var users []models.User
	if err := core.DB().Find(&users).Error; err != nil {
		return res.Error(err.Error(), 500)
	}
	return res.Success(users)
}

func (uc *UserController) Store(req *core.Request, res *core.Response) error {
	var body CreateUserRequest
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error())
	}

	user := models.User{Name: body.Name, Email: body.Email}
	if err := core.DB().Create(&user).Error; err != nil {
		return res.Error(err.Error(), 500)
	}
	return res.Success(user)
}

func (uc *UserController) Show(req *core.Request, res *core.Response) error {
	id := req.Input("id")
	if id == "" {
		return res.NotFound(core.GetLang().Trans("user.not_found"))
	}

	var user models.User
	if err := core.DB().First(&user, id).Error; err != nil {
		return res.NotFound(core.GetLang().Trans("user.not_found"))
	}
	return res.Success(user)
}
