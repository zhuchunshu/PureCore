package controllers

import (
	"os"

	middleware "purecore/app/Http/Middleware"
	models "purecore/app/Models"
	"purecore/core"
)

type AdminAuthController struct{}

type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=6"`
}

// Login authenticates an admin user and returns a JWT token
func (ac *AdminAuthController) Login(req *core.Request, res *core.Response) error {
	var body LoginRequest
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid credentials", 422)
	}

	var admin models.AdminUser
	if err := core.DB().Where("username = ?", body.Username).First(&admin).Error; err != nil {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	if !admin.CheckPassword(body.Password) {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	token, err := middleware.GenerateAdminToken(admin.ID, admin.Username)
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	return res.Success(map[string]interface{}{
		"token":    token,
		"username": admin.Username,
		"name":     admin.Name,
		"role":     admin.Role,
	})
}

// Profile returns the current admin user's profile from JWT token
func (ac *AdminAuthController) Profile(req *core.Request, res *core.Response) error {
	var admin models.AdminUser
	if err := core.DB().First(&admin, middleware.GetAdminUserID(req.Ctx())).Error; err != nil {
		return res.NotFound(core.GetLang().Trans("admin.user_not_found"))
	}

	return res.Success(map[string]interface{}{
		"id":       admin.ID,
		"username": admin.Username,
		"name":     admin.Name,
		"role":     admin.Role,
	})
}

// CheckAdminExists returns the count of admin users (public endpoint)
func (ac *AdminAuthController) CheckAdminExists(req *core.Request, res *core.Response) error {
	var count int64
	core.DB().Model(&models.AdminUser{}).Count(&count)
	return res.Success(map[string]interface{}{
		"exists": count > 0,
		"count":  count,
	})
}

// CreateAdmin creates a new admin user.
// First admin gets "super_admin" role; subsequent creations require authentication
// and assign the "admin" role.
func (ac *AdminAuthController) CreateAdmin(req *core.Request, res *core.Response) error {
	var body struct {
		Username string `json:"username" validate:"required,min=3"`
		Password string `json:"password" validate:"required,min=6"`
		Name     string `json:"name" validate:"required"`
	}
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error(), 422)
	}

	// Check how many admins exist
	var count int64
	core.DB().Model(&models.AdminUser{}).Count(&count)

	// First admin gets super_admin, subsequent require existing admin auth
	role := "admin"
	if count == 0 {
		role = "super_admin"
	} else {
		// If admins already exist, only allow authenticated admins to create
		adminID := middleware.GetAdminUserID(req.Ctx())
		if adminID == 0 {
			return res.Error(core.GetLang().Trans("admin.registration_disabled"), 403)
		}
	}

	admin := models.AdminUser{
		Username: body.Username,
		Name:     body.Name,
		Role:     role,
	}
	if err := admin.SetPassword(body.Password); err != nil {
		return res.Error(core.GetLang().Trans("admin.password_hash_failed"), 500)
	}

	if err := core.DB().Create(&admin).Error; err != nil {
		return res.Error(core.GetLang().Trans("admin.create_failed")+": "+err.Error(), 500)
	}

	token, _ := middleware.GenerateAdminToken(admin.ID, admin.Username)

	return res.Success(map[string]interface{}{
		"message":  core.GetLang().Trans("admin.register_success"),
		"token":    token,
		"username": admin.Username,
		"name":     admin.Name,
		"role":     admin.Role,
	})
}

// GetAdminRoutePrefix returns the admin route prefix from env
func GetAdminRoutePrefix() string {
	prefix := os.Getenv("ADMIN_ROUTE_PREFIX")
	if prefix == "" {
		prefix = "admin"
	}
	return "/api/v1/" + prefix
}
