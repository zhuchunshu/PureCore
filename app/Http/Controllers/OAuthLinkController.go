package controllers

import (
	"encoding/json"
	"strings"
	"time"

	middleware "purecore/app/Http/Middleware"
	models "purecore/app/Models"
	"purecore/core"
	"purecore/core/oauth"
)

// OAuthLinkController handles OAuth account linking after a callback
// when the OAuth identity is not yet linked to any PureCore user.
// It provides two endpoints:
//   - LinkRegister: create a new user account and bind the OAuth identity
//   - LinkLogin: login to an existing account and bind the OAuth identity
type OAuthLinkController struct{}

// LinkRegisterRequest is the request body for OAuth link + register.
type LinkRegisterRequest struct {
	LinkToken      string `json:"link_token" validate:"required"`
	Name           string `json:"name" validate:"required,min=2"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	TurnstileToken string `json:"turnstile_token"`
}

// LinkLoginRequest is the request body for OAuth link + login.
type LinkLoginRequest struct {
	LinkToken      string `json:"link_token" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	TurnstileToken string `json:"turnstile_token"`
}

// LinkRegister creates a new user account and links the OAuth identity to it.
// POST /api/v1/oauth/:provider/link/register
func (c *OAuthLinkController) LinkRegister(req *core.Request, res *core.Response) error {
	var body LinkRegisterRequest
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid request: "+err.Error(), 422)
	}

	// Verify Turnstile if enabled for public login
	if core.IsTurnstileEnabled("turnstile_public_login") {
		if err := core.VerifyTurnstile(body.TurnstileToken); err != nil {
			return res.Error("Captcha verification failed: "+err.Error(), 422)
		}
	}

	// Parse and validate the link token
	oAuthInfo, err := oauth.ParseLinkToken(body.LinkToken)
	if err != nil {
		return res.Error("Invalid or expired link token", 422)
	}

	// Check if email already exists (including soft-deleted records)
	var existing models.User
	if err := core.DB().Unscoped().Where("email = ?", body.Email).First(&existing).Error; err == nil {
		if existing.DeletedAt.Valid {
			// Hard-delete the soft-deleted record so the unique index won't conflict
			core.DB().Unscoped().Delete(&existing)
		} else {
			return res.Error(core.GetLang().Trans("user.email_exists"), 409)
		}
	}

	// Create user with real password (not a random placeholder)
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

	// Check if OAuth provider_id already exists (upsert)
	now := time.Now()
	var existingOAuth models.OAuthAccount
	if err := core.DB().Where("provider = ? AND provider_id = ?", oAuthInfo.Provider, oAuthInfo.ProviderID).First(&existingOAuth).Error; err == nil {
		// Record exists — update it to point to the new user
		core.DB().Model(&existingOAuth).Updates(map[string]interface{}{
			"user_id":      user.ID,
			"email":        user.Email,
			"name":         user.Name,
			"avatar_url":   oAuthInfo.AvatarURL,
			"access_token": oAuthInfo.AccessToken,
			"token_expiry": oAuthInfo.TokenExpiry,
			"raw_data":     rawDataJSON(oAuthInfo),
			"updated_at":   now,
		})
	} else {
		// No existing record — create new one
		oauthAccount := models.OAuthAccount{
			UserID:      user.ID,
			Provider:    oAuthInfo.Provider,
			ProviderID:  oAuthInfo.ProviderID,
			Email:       user.Email,
			Name:        user.Name,
			AvatarURL:   oAuthInfo.AvatarURL,
			AccessToken: oAuthInfo.AccessToken,
			TokenExpiry: oAuthInfo.TokenExpiry,
		}
		if oAuthInfo.RawData != nil {
			rawBytes, _ := json.Marshal(oAuthInfo.RawData)
			oauthAccount.RawData = string(rawBytes)
		}
		if err := core.DB().Create(&oauthAccount).Error; err != nil {
			core.DB().Unscoped().Delete(&user)
			return res.Error("Failed to link OAuth account: "+err.Error(), 500)
		}
	}

	// Update last login
	core.DB().Model(&user).Update("last_login_at", now)

	// Login the user and return tokens
	return c.loginUser(req, res, &user)
}

// LinkLogin authenticates an existing user and links the OAuth identity to them.
// POST /api/v1/oauth/:provider/link/login
func (c *OAuthLinkController) LinkLogin(req *core.Request, res *core.Response) error {
	var body LinkLoginRequest
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid credentials", 422)
	}

	// Verify Turnstile if enabled for public login
	if core.IsTurnstileEnabled("turnstile_public_login") {
		if err := core.VerifyTurnstile(body.TurnstileToken); err != nil {
			return res.Error("Captcha verification failed: "+err.Error(), 422)
		}
	}

	// Parse and validate the link token
	oAuthInfo, err := oauth.ParseLinkToken(body.LinkToken)
	if err != nil {
		return res.Error("Invalid or expired link token", 422)
	}

	// Authenticate user by email and password
	var user models.User
	if err := core.DB().Where("email = ?", body.Email).First(&user).Error; err != nil {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	if !user.CheckPassword(body.Password) {
		return res.Error(core.GetLang().Trans("admin.invalid_credentials"), 401)
	}

	// Check if this OAuth provider_id is already linked to another user
	var existingOAuth models.OAuthAccount
	if err := core.DB().Where("provider = ? AND provider_id = ?", oAuthInfo.Provider, oAuthInfo.ProviderID).First(&existingOAuth).Error; err == nil {
		if existingOAuth.UserID != 0 {
			if existingOAuth.UserID != user.ID {
				return res.Error("This OAuth account is already linked to another user", 409)
			}
			// Already linked to this user — just login
			return c.loginUser(req, res, &user)
		}
		// UserID == 0: dirty placeholder record from previous flow — update it
		now := time.Now()
		core.DB().Model(&existingOAuth).Updates(map[string]interface{}{
			"user_id":      user.ID,
			"email":        user.Email,
			"name":         user.Name,
			"avatar_url":   oAuthInfo.AvatarURL,
			"access_token": oAuthInfo.AccessToken,
			"token_expiry": oAuthInfo.TokenExpiry,
			"raw_data":     rawDataJSON(oAuthInfo),
			"updated_at":   now,
		})
		// Update user's avatar if not set
		if user.Avatar == "" && oAuthInfo.AvatarURL != "" {
			core.DB().Model(&user).Update("avatar", oAuthInfo.AvatarURL)
		}
		core.DB().Model(&user).Update("last_login_at", now)
		return c.loginUser(req, res, &user)
	}

	// Create the OAuth account link
	now := time.Now()
	oauthAccount := models.OAuthAccount{
		UserID:      user.ID,
		Provider:    oAuthInfo.Provider,
		ProviderID:  oAuthInfo.ProviderID,
		Email:       user.Email,
		Name:        user.Name,
		AvatarURL:   oAuthInfo.AvatarURL,
		AccessToken: oAuthInfo.AccessToken,
		TokenExpiry: oAuthInfo.TokenExpiry,
	}
	if oAuthInfo.RawData != nil {
		rawBytes, _ := json.Marshal(oAuthInfo.RawData)
		oauthAccount.RawData = string(rawBytes)
	}
	if err := core.DB().Create(&oauthAccount).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			// Race condition: another request already linked it — update instead
			core.DB().Model(&models.OAuthAccount{}).
				Where("provider = ? AND provider_id = ?", oAuthInfo.Provider, oAuthInfo.ProviderID).
				Updates(map[string]interface{}{
					"user_id":      user.ID,
					"email":        user.Email,
					"name":         user.Name,
					"avatar_url":   oAuthInfo.AvatarURL,
					"access_token": oAuthInfo.AccessToken,
					"token_expiry": oAuthInfo.TokenExpiry,
					"raw_data":     rawDataJSON(oAuthInfo),
					"updated_at":   now,
				})
		} else {
			return res.Error("Failed to link OAuth account: "+err.Error(), 500)
		}
	}

	// Update user's avatar if not set
	if user.Avatar == "" && oAuthInfo.AvatarURL != "" {
		core.DB().Model(&user).Update("avatar", oAuthInfo.AvatarURL)
	}

	// Update last login time
	core.DB().Model(&user).Update("last_login_at", now)

	// Login the user and return tokens
	return c.loginUser(req, res, &user)
}

// loginUser generates tokens and creates a session for a user, returning JSON.
// Mirrors OAuthController.loginUser() for consistency.
func (c *OAuthLinkController) loginUser(req *core.Request, res *core.Response, user *models.User) error {
	accessToken, err := middleware.GenerateUserToken(user.ID, user.Name)
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		return res.Error(core.GetLang().Trans("admin.token_generate_failed"), 500)
	}

	now := time.Now()
	core.DB().Model(user).Updates(map[string]interface{}{
		"refresh_token": refreshToken,
		"last_login_at": now,
	})

	// Mark all previous sessions as not current
	core.DB().Model(&models.UserSession{}).Where("user_id = ?", user.ID).Updates(map[string]interface{}{"is_current": false})
	// Create a new session record
	CreateSession(req.Ctx(), user.ID)

	return res.Success(map[string]interface{}{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"name":          user.Name,
		"email":         user.Email,
		"linked":        true,
	})
}

// rawDataJSON returns the RawData from LinkTokenData as a JSON string,
// or empty string if RawData is nil.
func rawDataJSON(info *oauth.LinkTokenData) string {
	if info == nil || info.RawData == nil {
		return ""
	}
	rawBytes, err := json.Marshal(info.RawData)
	if err != nil {
		return ""
	}
	return string(rawBytes)
}
