package controllers

import (
	middleware "purecore/app/Http/Middleware"
	models "purecore/app/Models"
	"purecore/core"
	"time"

	"github.com/gofiber/fiber/v3"
)

type UserAuthController struct{}

// RegisterRequest is the request body for user registration
type RegisterRequest struct {
	Name           string `json:"name" validate:"required,min=2"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	TurnstileToken string `json:"turnstile_token"`
}

// LoginRequest is the request body for user login
type UserLoginRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	TurnstileToken string `json:"turnstile_token"`
}

// Register creates a new regular user account
func (uc *UserAuthController) Register(req *core.Request, res *core.Response) error {
	var body RegisterRequest
	if err := req.Validate(&body); err != nil {
		return res.Error(err.Error(), 422)
	}

	// Verify Turnstile if enabled for public login
	if core.IsTurnstileEnabled("turnstile_public_login") {
		if err := core.VerifyTurnstile(body.TurnstileToken); err != nil {
			return res.Error("Captcha verification failed: "+err.Error(), 422)
		}
	}

	// Check if email already exists
	var existing models.User
	if err := core.DB().Where("email = ?", body.Email).First(&existing).Error; err == nil {
		return res.Error(core.GetLang().Trans("user.email_exists"), 409)
	}

	user := models.User{
		Name:  body.Name,
		Email: body.Email,
	}
	if err := user.SetPassword(body.Password); err != nil {
		return res.Error(core.GetLang().Trans("admin.password_hash_failed"), 500)
	}

	if err := core.DB().Create(&user).Error; err != nil {
		return res.Error(core.GetLang().Trans("admin.create_failed")+": "+err.Error(), 500)
	}

	accessToken, err := middleware.GenerateUserToken(user.ID, user.Name)
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	// Save the refresh token in the database
	core.DB().Model(&user).Update("refresh_token", refreshToken)

	return res.Success(map[string]interface{}{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"name":          user.Name,
		"email":         user.Email,
	})
}

// Login authenticates a regular user and returns JWT tokens
func (uc *UserAuthController) Login(req *core.Request, res *core.Response) error {
	var body UserLoginRequest
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid credentials", 422)
	}

	// Verify Turnstile if enabled for public login
	if core.IsTurnstileEnabled("turnstile_public_login") {
		if err := core.VerifyTurnstile(body.TurnstileToken); err != nil {
			return res.Error("Captcha verification failed: "+err.Error(), 422)
		}
	}

	var user models.User
	if err := core.DB().Where("email = ?", body.Email).First(&user).Error; err != nil {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	if !user.CheckPassword(body.Password) {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	accessToken, err := middleware.GenerateUserToken(user.ID, user.Name)
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	// Save the refresh token and update last login time
	now := time.Now()
	core.DB().Model(&user).Updates(map[string]interface{}{
		"refresh_token": refreshToken,
		"last_login_at": now,
	})

	return res.Success(map[string]interface{}{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"name":          user.Name,
		"email":         user.Email,
	})
}

// Profile returns the current authenticated user's profile
func (uc *UserAuthController) Profile(req *core.Request, res *core.Response) error {
	userID := getUserID(req.Ctx())
	if userID == 0 {
		return res.Unauthorized()
	}

	var user models.User
	if err := core.DB().First(&user, userID).Error; err != nil {
		return res.NotFound(core.GetLang().Trans("user.not_found"))
	}

	return res.Success(map[string]interface{}{
		"id":                user.ID,
		"name":              user.Name,
		"email":             user.Email,
		"avatar":            user.Avatar,
		"bio":               user.Bio,
		"status":            user.Status,
		"email_verified_at": user.EmailVerifiedAt,
		"last_login_at":     user.LastLoginAt,
		"created_at":        user.CreatedAt,
		"updated_at":        user.UpdatedAt,
	})
}

// Refresh generates a new access token using a valid refresh token
func (uc *UserAuthController) Refresh(req *core.Request, res *core.Response) error {
	var body struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid request", 422)
	}

	// Find the user with this refresh token
	var user models.User
	if err := core.DB().Where("refresh_token = ?", body.RefreshToken).First(&user).Error; err != nil {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	// Generate new access token
	accessToken, err := middleware.GenerateUserToken(user.ID, user.Name)
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	// Rotate the refresh token for security
	newRefreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}
	core.DB().Model(&user).Update("refresh_token", newRefreshToken)

	return res.Success(map[string]interface{}{
		"token":         accessToken,
		"refresh_token": newRefreshToken,
	})
}

// getUserID extracts the user ID from the JWT claims stored in Locals by the Auth middleware
func getUserID(c fiber.Ctx) uint {
	userVal := c.Locals("user")
	if userVal == nil {
		return 0
	}
	userMap, ok := userVal.(map[string]string)
	if !ok {
		return 0
	}
	// The Auth middleware stores user ID as a string in the map
	idStr := userMap["id"]
	if idStr == "" {
		return 0
	}
	// Parse the string back to uint - we need to handle this carefully
	var id uint
	for _, ch := range idStr {
		if ch < '0' || ch > '9' {
			return 0
		}
		id = id*10 + uint(ch-'0')
	}
	return id
}
