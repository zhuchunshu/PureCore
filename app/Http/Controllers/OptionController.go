package controllers

import (
	"purecore/core"
)

type OptionController struct{}

type UpdateOptionRequest struct {
	Name  string `json:"name" validate:"required,min=1"`
	Value string `json:"value"`
}

type BatchUpdateOptionsRequest struct {
	Options map[string]string `json:"options" validate:"required,min=1"`
}

// GetAll returns all web options as key-value pairs
func (c *OptionController) GetAll(req *core.Request, res *core.Response) error {
	options := make(map[string]string)
	db := core.DB()
	var rows []struct {
		Name  string
		Value string
	}
	if err := db.Table("web_options").Select("name", "value").Find(&rows).Error; err != nil {
		return res.Error(err.Error(), 500)
	}
	for _, row := range rows {
		options[row.Name] = row.Value
	}
	return res.Success(options)
}

// Set updates or creates a web option
func (c *OptionController) Set(req *core.Request, res *core.Response) error {
	var body UpdateOptionRequest
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error(), 422)
	}

	if err := core.AdminOptionSet(body.Name, body.Value); err != nil {
		return res.Error(err.Error(), 500)
	}
	return res.Success(map[string]string{
		"name":  body.Name,
		"value": body.Value,
	})
}

// BatchSet updates or creates multiple web options in a single transaction
func (c *OptionController) BatchSet(req *core.Request, res *core.Response) error {
	var body BatchUpdateOptionsRequest
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error(), 422)
	}

	if len(body.Options) == 0 {
		return res.Error("options cannot be empty", 422)
	}

	if err := core.AdminOptionSetMany(body.Options); err != nil {
		return res.Error(err.Error(), 500)
	}
	return res.Success(map[string]interface{}{
		"count":   len(body.Options),
		"options": body.Options,
	})
}
