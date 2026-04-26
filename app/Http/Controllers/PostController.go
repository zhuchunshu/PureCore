package controllers

import (
	models "purecore/app/Models"
	"purecore/core"
)

type PostController struct{}

type CreatePostRequest struct {
	Name string `json:"name" validate:"required,min=2"`
}

func (c *PostController) Index(req *core.Request, res *core.Response) error {
	var records []models.Post
	if err := core.DB().Find(&records).Error; err != nil {
		return res.Error(err.Error(), 500)
	}
	return res.Success(records)
}

func (c *PostController) Store(req *core.Request, res *core.Response) error {
	var body CreatePostRequest
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error())
	}

	record := models.Post{Name: body.Name}
	if err := core.DB().Create(&record).Error; err != nil {
		return res.Error(err.Error(), 500)
	}
	return res.Success(record)
}

func (c *PostController) Show(req *core.Request, res *core.Response) error {
	id := req.Input("id")
	if id == "" {
		return res.NotFound(core.GetLang().Trans("db.record_not_found"))
	}

	var record models.Post
	if err := core.DB().First(&record, id).Error; err != nil {
		return res.NotFound(core.GetLang().Trans("db.record_not_found"))
	}
	return res.Success(record)
}
