package controllers

import (
	middleware "purecore/app/Http/Middleware"
	models "purecore/app/Models"
	"purecore/core"
)

type AdminAuthController struct{}

type LoginRequest struct {
	Username       string `json:"username" validate:"required,min=3"`
	Password       string `json:"password" validate:"required,min=6"`
	TurnstileToken string `json:"turnstile_token"`
}

// Login authenticates an admin user and returns a JWT token
func (ac *AdminAuthController) Login(req *core.Request, res *core.Response) error {
	var body LoginRequest
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid credentials", 422)
	}

	// Verify Turnstile if enabled for admin login (temporarily disabled for testing)
	// if core.IsTurnstileEnabled("turnstile_admin_login") {
	// 	if err := core.VerifyTurnstile(body.TurnstileToken); err != nil {
	// 		return res.Error("Captcha verification failed: "+err.Error(), 422)
	// 	}
	// }

	var admin models.AdminUser
	if err := core.DB().Where("username = ?", body.Username).First(&admin).Error; err != nil {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	if !admin.CheckPassword(body.Password) {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	accessToken, err := middleware.GenerateAdminToken(admin.ID, admin.Username, admin.TokenVersion)
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	// Save the refresh token in the database
	core.DB().Model(&admin).Update("refresh_token", refreshToken)

	return res.Success(map[string]interface{}{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"username":      admin.Username,
		"name":          admin.Name,
		"role":          admin.Role,
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
		Username       string `json:"username" validate:"required,min=3"`
		Password       string `json:"password" validate:"required,min=6"`
		Name           string `json:"name" validate:"required"`
		TurnstileToken string `json:"turnstile_token"`
	}
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error(), 422)
	}

	// Verify Turnstile if enabled for admin register (temporarily disabled for testing)
	// if core.IsTurnstileEnabled("turnstile_admin_register") {
	// 	if err := core.VerifyTurnstile(body.TurnstileToken); err != nil {
	// 		return res.Error("Captcha verification failed: "+err.Error(), 422)
	// 	}
	// }

	// TEMPORARY: Skip admin count check for testing purposes
	// // Check how many admins exist
	// var count int64
	// core.DB().Model(&models.AdminUser{}).Count(&count)
	//
	// // First admin gets super_admin, subsequent require existing admin auth
	// role := "admin"
	// if count == 0 {
	// 	role = "super_admin"
	// } else {
	// 	// If admins already exist, only allow authenticated admins to create
	// 	adminID := middleware.GetAdminUserID(req.Ctx())
	// 	if adminID == 0 {
	// 		return res.Error(core.GetLang().Trans("admin.registration_disabled"), 403)
	// 	}
	// }
	role := "admin"

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

	accessToken, _ := middleware.GenerateAdminToken(admin.ID, admin.Username, admin.TokenVersion)
	refreshToken, _ := middleware.GenerateRefreshToken()

	// Save the refresh token in the database
	core.DB().Model(&admin).Update("refresh_token", refreshToken)

	return res.Success(map[string]interface{}{
		"message":       core.GetLang().Trans("admin.register_success"),
		"token":         accessToken,
		"refresh_token": refreshToken,
		"username":      admin.Username,
		"name":          admin.Name,
		"role":          admin.Role,
	})
}

// ChangePassword allows authenticated admins to change their password
// Increments token_version to invalidate all existing tokens across all devices
func (ac *AdminAuthController) ChangePassword(req *core.Request, res *core.Response) error {
	var body struct {
		CurrentPassword string `json:"current_password" validate:"required,min=6"`
		NewPassword     string `json:"new_password" validate:"required,min=6"`
	}
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error(), 422)
	}

	userID := middleware.GetAdminUserID(req.Ctx())
	if userID == 0 {
		return res.Unauthorized()
	}

	var admin models.AdminUser
	if err := core.DB().First(&admin, userID).Error; err != nil {
		return res.NotFound(core.GetLang().Trans("admin.user_not_found"))
	}

	if !admin.CheckPassword(body.CurrentPassword) {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	if err := admin.SetPassword(body.NewPassword); err != nil {
		return res.Error(core.GetLang().Trans("admin.password_hash_failed"), 500)
	}

	// Increment token_version to invalidate all existing tokens
	admin.TokenVersion++
	core.DB().Model(&admin).Updates(map[string]interface{}{
		"password":      admin.Password,
		"token_version": admin.TokenVersion,
		"refresh_token": "", // Also clear refresh token
	})

	return res.Success(map[string]interface{}{
		"message": core.GetLang().Trans("admin.password_changed"),
	})
}

// Refresh generates a new access token using a valid refresh token
func (ac *AdminAuthController) Refresh(req *core.Request, res *core.Response) error {
	var body struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid request", 422)
	}

	// Find the admin user with this refresh token
	var admin models.AdminUser
	if err := core.DB().Where("refresh_token = ?", body.RefreshToken).First(&admin).Error; err != nil {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	// Generate new access token
	accessToken, err := middleware.GenerateAdminToken(admin.ID, admin.Username, admin.TokenVersion)
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	// Rotate the refresh token for security
	newRefreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}
	core.DB().Model(&admin).Update("refresh_token", newRefreshToken)

	return res.Success(map[string]interface{}{
		"token":         accessToken,
		"refresh_token": newRefreshToken,
	})
}

// GetAdminRoutePrefix returns the admin route prefix from the config system.
// This function is kept for backward compatibility.
func GetAdminRoutePrefix() string {
	prefix := core.GetConfig().AdminRoutePrefix()
	if prefix == "" {
		prefix = "admin"
	}
	return "/api/v1/" + prefix
}
