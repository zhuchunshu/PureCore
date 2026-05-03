package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	middleware "purecore/app/Http/Middleware"
	models "purecore/app/Models"
	"purecore/core"
	"purecore/core/oauth"

	"github.com/gofiber/fiber/v3"
)

// OAuthController handles third-party OAuth login flows.
type OAuthController struct{}

// Providers lists all registered OAuth providers with their display info and enabled status.
func (c *OAuthController) Providers(req *core.Request, res *core.Response) error {
	providers := oauth.All()
	list := make([]map[string]interface{}, 0, len(providers))
	for _, p := range providers {
		enabled := core.AdminOption(oauth.FullConfigKey(p, "enabled")) == "true"
		loginEnabled := core.AdminOption(oauth.FullConfigKey(p, "login_enabled")) == "true"
		registerEnabled := core.AdminOption(oauth.FullConfigKey(p, "register_enabled")) == "true"
		list = append(list, map[string]interface{}{
			"name":             p.Name(),
			"display_name":     p.DisplayName(),
			"enabled":          enabled,
			"login_enabled":    loginEnabled,
			"register_enabled": registerEnabled,
		})
	}
	return res.Success(list)
}

// Authorize generates the OAuth authorization URL for a provider and returns it.
// Request body: { "redirect": "/dashboard" } (optional, defaults to "/")
func (c *OAuthController) Authorize(req *core.Request, res *core.Response) error {
	providerName := req.Ctx().Params("provider")
	provider := oauth.Get(providerName)
	if provider == nil {
		return res.Error("Unsupported OAuth provider", 400)
	}

	type AuthRequest struct {
		Redirect string `json:"redirect"`
	}
	var body AuthRequest
	if err := req.Validate(&body); err != nil {
		body.Redirect = "/"
	}

	// Default action is "login"
	action := "login"

	state, err := oauth.GenerateState(providerName, body.Redirect, action)
	if err != nil {
		return res.Error("Failed to generate state", 500)
	}

	authURL := provider.GetAuthURL(state)
	return res.Success(map[string]interface{}{
		"url": authURL,
	})
}

// Callback handles the OAuth provider redirect after user authorization.
// This is the endpoint the provider redirects to after user consent.
// It exchanges the code, fetches user info, and redirects to the frontend
// OAuth callback page with the appropriate parameters (link_token, status, etc.)
func (c *OAuthController) Callback(req *core.Request, res *core.Response) error {
	providerName := req.Ctx().Params("provider")
	provider := oauth.Get(providerName)
	if provider == nil {
		return res.Error("Unsupported OAuth provider", 400)
	}

	code := req.Ctx().Query("code")
	stateStr := req.Ctx().Query("state")
	if code == "" || stateStr == "" {
		return res.Error("Missing code or state", 422)
	}

	// Parse state token
	state, err := oauth.ParseState(stateStr)
	if err != nil {
		return res.Error("Invalid or expired state token", 422)
	}

	// Exchange code for user info
	exchangeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	userInfo, err := provider.Exchange(exchangeCtx, code)
	if err != nil {
		return res.Error("Failed to exchange OAuth code: "+err.Error(), 500)
	}

	// Check if this OAuth account is already linked to a user
	var existingOAuth models.OAuthAccount
	alreadyLinked := false
	var linkedUserID uint
	if err := core.DB().Where("provider = ? AND provider_id = ?", providerName, userInfo.ProviderID).First(&existingOAuth).Error; err == nil {
		alreadyLinked = true
		linkedUserID = existingOAuth.UserID
	}

	// Base redirect URL on the frontend
	frontendBase := getFrontendURL(req)
	originalRedirect := state.Redirect
	if originalRedirect == "" {
		originalRedirect = "/"
	}

	// Case 1: Already linked → log the user in and redirect home
	if alreadyLinked {
		var user models.User
		if err := core.DB().First(&user, linkedUserID).Error; err != nil {
			return res.Error("Linked user not found", 500)
		}
		c.loginUser(req, res, &user)
		// loginUser writes JSON, but we need to redirect instead. We'll override.
		// Remove the JSON body already written and redirect.
		// Actually loginUser writes JSON to res, but we need a redirect.
		// So we do login logic inline and set tokens via cookies, then redirect.
		return c.loginWithRedirect(req, res, &user, frontendBase+"/oauth/callback?status=linked&redirect="+urlQueryEscape(originalRedirect))
	}

	// Check if there's a logged-in user (optional)
	currentUserID := getUserID(req.Ctx())

	// Case 2: Logged-in user → bind automatically
	if currentUserID != 0 {
		err := c.bindOAuthSilent(currentUserID, providerName, userInfo)
		if err != nil {
			return res.Error("Failed to bind OAuth: "+err.Error(), 500)
		}
		redirectURL := frontendBase + "/oauth/callback?status=bound&redirect=" + urlQueryEscape(originalRedirect)
		return req.Ctx().Redirect().To(redirectURL)
	}

	// Case 3: Not linked and not logged in → generate link token and redirect to OAuth callback page
	linkToken, err := oauth.GenerateLinkToken(providerName, userInfo)
	if err != nil {
		return res.Error("Failed to generate link token", 500)
	}

	// Upsert OAuth account record with latest info
	core.DB().Where(models.OAuthAccount{
		Provider:   providerName,
		ProviderID: userInfo.ProviderID,
	}).Assign(models.OAuthAccount{
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		AvatarURL: userInfo.AvatarURL,
	}).FirstOrCreate(&models.OAuthAccount{})

	redirectURL := frontendBase + "/oauth/callback?status=unlinked&link_token=" + urlQueryEscape(linkToken) +
		"&provider=" + urlQueryEscape(providerName) +
		"&email=" + urlQueryEscape(userInfo.Email) +
		"&name=" + urlQueryEscape(userInfo.Name) +
		"&avatar_url=" + urlQueryEscape(userInfo.AvatarURL) +
		"&redirect=" + urlQueryEscape(originalRedirect)
	return req.Ctx().Redirect().To(redirectURL)
}

// Register creates a new user account and binds it to the OAuth provider.
// Requires a valid link_token from the callback.
func (c *OAuthController) Register(req *core.Request, res *core.Response) error {
	type RegisterOAuthRequest struct {
		LinkToken string `json:"link_token" validate:"required"`
		Name      string `json:"name" validate:"required,min=2"`
		Email     string `json:"email" validate:"required,email"`
	}
	var body RegisterOAuthRequest
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid request: "+err.Error(), 422)
	}

	// Parse link token
	oAuthInfo, err := oauth.ParseLinkToken(body.LinkToken)
	if err != nil {
		return res.Error("Invalid or expired link token", 422)
	}

	// Check if email already exists
	var existing models.User
	if err := core.DB().Where("email = ?", body.Email).First(&existing).Error; err == nil {
		return res.Error(core.GetLang().Trans("user.email_exists"), 409)
	}

	// Create user (no password for OAuth-only users; can set password later)
	user := models.User{
		Name:  body.Name,
		Email: body.Email,
	}
	// Set a random password placeholder so the account can't be logged into with password
	_ = user.SetPassword(fmt.Sprintf("oauth-%s-%d", oAuthInfo.Provider, time.Now().UnixNano()))

	if err := core.DB().Create(&user).Error; err != nil {
		return res.Error(core.GetLang().Trans("admin.create_failed")+": "+err.Error(), 500)
	}

	// Create OAuth account link
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
		return res.Error("Failed to link OAuth account: "+err.Error(), 500)
	}

	// Update last login
	core.DB().Model(&user).Update("last_login_at", now)

	// Login the user
	return c.loginUser(req, res, &user)
}

// Bind links an OAuth provider to the currently logged-in user.
// Requires auth middleware.
func (c *OAuthController) Bind(req *core.Request, res *core.Response) error {
	type BindOAuthRequest struct {
		LinkToken string `json:"link_token" validate:"required"`
	}
	var body BindOAuthRequest
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid request: "+err.Error(), 422)
	}

	userID := getUserID(req.Ctx())
	if userID == 0 {
		return res.Unauthorized()
	}

	oAuthInfo, err := oauth.ParseLinkToken(body.LinkToken)
	if err != nil {
		return res.Error("Invalid or expired link token", 422)
	}

	// Check if this provider ID is already linked to another user
	var existing models.OAuthAccount
	if err := core.DB().Where("provider = ? AND provider_id = ?", oAuthInfo.Provider, oAuthInfo.ProviderID).First(&existing).Error; err == nil {
		if existing.UserID != userID {
			return res.Error("This OAuth account is already linked to another user", 409)
		}
		return res.Success("Already linked")
	}

	// Create the link
	now := time.Now()
	oauthAccount := models.OAuthAccount{
		UserID:      userID,
		Provider:    oAuthInfo.Provider,
		ProviderID:  oAuthInfo.ProviderID,
		Email:       oAuthInfo.Email,
		Name:        oAuthInfo.Name,
		AvatarURL:   oAuthInfo.AvatarURL,
		AccessToken: oAuthInfo.AccessToken,
		TokenExpiry: oAuthInfo.TokenExpiry,
	}
	if oAuthInfo.RawData != nil {
		rawBytes, _ := json.Marshal(oAuthInfo.RawData)
		oauthAccount.RawData = string(rawBytes)
	}
	if err := core.DB().Create(&oauthAccount).Error; err != nil {
		return res.Error("Failed to link OAuth account: "+err.Error(), 500)
	}

	// Update user's avatar if not set
	var user models.User
	if err := core.DB().First(&user, userID).Error; err == nil {
		if user.Avatar == "" && oAuthInfo.AvatarURL != "" {
			core.DB().Model(&user).Update("avatar", oAuthInfo.AvatarURL)
		}
	}

	_ = now // Suppress unused warning
	return res.Success(map[string]interface{}{
		"message": "OAuth account linked successfully",
	})
}

// Accounts returns all OAuth accounts linked to the current user.
func (c *OAuthController) Accounts(req *core.Request, res *core.Response) error {
	userID := getUserID(req.Ctx())
	if userID == 0 {
		return res.Unauthorized()
	}

	var accounts []models.OAuthAccount
	core.DB().Where("user_id = ?", userID).Find(&accounts)
	return res.Success(accounts)
}

// Unlink removes an OAuth account link from the current user.
func (c *OAuthController) Unlink(req *core.Request, res *core.Response) error {
	userID := getUserID(req.Ctx())
	if userID == 0 {
		return res.Unauthorized()
	}

	accountID := req.Ctx().Params("id")
	var account models.OAuthAccount
	if err := core.DB().Where("id = ? AND user_id = ?", accountID, userID).First(&account).Error; err != nil {
		return res.NotFound("OAuth account not found")
	}

	if err := core.DB().Delete(&account).Error; err != nil {
		return res.Error("Failed to unlink OAuth account", 500)
	}

	return res.Success("OAuth account unlinked")
}

// AdminGetSettings returns all OAuth provider configurations for the admin panel.
func (c *OAuthController) AdminGetSettings(req *core.Request, res *core.Response) error {
	providers := oauth.All()
	settings := make(map[string]map[string]string)
	for _, p := range providers {
		prefix := p.ConfigKeyPrefix()
		keys := []string{"enabled", "login_enabled", "register_enabled", "client_id", "client_secret", "redirect_url"}
		providerSettings := make(map[string]string)
		for _, key := range keys {
			fullKey := prefix + "_" + key
			val := core.AdminOption(fullKey, "")
			providerSettings[key] = val
		}
		providerSettings["display_name"] = p.DisplayName()
		providerSettings["name"] = p.Name()
		settings[p.Name()] = providerSettings
	}
	return res.Success(settings)
}

// AdminSetSettings updates OAuth provider configurations in bulk.
func (c *OAuthController) AdminSetSettings(req *core.Request, res *core.Response) error {
	type ProviderSettingsUpdate struct {
		Provider string            `json:"provider" validate:"required"`
		Settings map[string]string `json:"settings"`
	}
	type BulkUpdateRequest struct {
		Providers []ProviderSettingsUpdate `json:"providers"`
	}
	var body BulkUpdateRequest
	if err := req.Validate(&body); err != nil {
		return res.Error("Invalid request: "+err.Error(), 422)
	}

	options := make(map[string]string)
	for _, ps := range body.Providers {
		provider := oauth.Get(ps.Provider)
		if provider == nil {
			continue
		}
		for key, value := range ps.Settings {
			fullKey := provider.ConfigKeyPrefix() + "_" + key
			options[fullKey] = value
		}
	}

	if err := core.AdminOptionSetMany(options); err != nil {
		return res.Error("Failed to save settings: "+err.Error(), 500)
	}
	return res.Success("Settings saved")
}

// loginUser generates tokens and creates a session for a user, returning JSON.
func (c *OAuthController) loginUser(req *core.Request, res *core.Response, user *models.User) error {
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

// loginWithRedirect logs the user in and sets tokens as non-HttpOnly cookies,
// then redirects to the given URL.
func (c *OAuthController) loginWithRedirect(req *core.Request, res *core.Response, user *models.User, redirectTo string) error {
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

	core.DB().Model(&models.UserSession{}).Where("user_id = ?", user.ID).Updates(map[string]interface{}{"is_current": false})
	CreateSession(req.Ctx(), user.ID)

	// Set tokens as cookies (not HttpOnly so JS can read them)
	req.Ctx().Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Secure:   false,
		HTTPOnly: false,
		SameSite: "Lax",
	})
	req.Ctx().Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Secure:   false,
		HTTPOnly: false,
		SameSite: "Lax",
	})

	return req.Ctx().Redirect().To(redirectTo)
}

// bindOAuthSilent links an OAuth account to the given user ID silently (no response).
func (c *OAuthController) bindOAuthSilent(userID uint, providerName string, userInfo *oauth.UserInfo) error {
	now := time.Now()
	oauthAccount := models.OAuthAccount{
		UserID:     userID,
		Provider:   providerName,
		ProviderID: userInfo.ProviderID,
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		AvatarURL:  userInfo.AvatarURL,
	}
	if userInfo.Raw != nil {
		rawBytes, _ := json.Marshal(userInfo.Raw)
		oauthAccount.RawData = string(rawBytes)
	}
	if err := core.DB().Create(&oauthAccount).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			core.DB().Model(&models.OAuthAccount{}).
				Where("provider = ? AND provider_id = ?", providerName, userInfo.ProviderID).
				Updates(map[string]interface{}{
					"email":      userInfo.Email,
					"name":       userInfo.Name,
					"avatar_url": userInfo.AvatarURL,
					"updated_at": now,
				})
		} else {
			return err
		}
	}

	var user models.User
	if err := core.DB().First(&user, userID).Error; err == nil {
		if user.Avatar == "" && userInfo.AvatarURL != "" {
			core.DB().Model(&user).Update("avatar", userInfo.AvatarURL)
		}
	}
	return nil
}

// ---------- helpers ----------

// getFrontendURL returns the base URL of the frontend (no trailing slash).
func getFrontendURL(req *core.Request) string {
	// Use the request's own scheme and host; the frontend is served on the same origin
	// in both dev (Vite proxy) and production (Go serves static files).
	scheme := req.Ctx().Protocol()
	host := req.Ctx().Hostname()
	return scheme + "://" + host
}

func urlQueryEscape(s string) string {
	return url.QueryEscape(s)
}

// bindOAuth links an OAuth account to the currently logged-in user.
func (c *OAuthController) bindOAuth(req *core.Request, res *core.Response, userID uint, providerName string, userInfo *oauth.UserInfo) error {
	now := time.Now()
	oauthAccount := models.OAuthAccount{
		UserID:     userID,
		Provider:   providerName,
		ProviderID: userInfo.ProviderID,
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		AvatarURL:  userInfo.AvatarURL,
	}
	if userInfo.Raw != nil {
		rawBytes, _ := json.Marshal(userInfo.Raw)
		oauthAccount.RawData = string(rawBytes)
	}
	if err := core.DB().Create(&oauthAccount).Error; err != nil {
		// If already exists (unique constraint), just update
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			core.DB().Model(&models.OAuthAccount{}).
				Where("provider = ? AND provider_id = ?", providerName, userInfo.ProviderID).
				Updates(map[string]interface{}{
					"email":      userInfo.Email,
					"name":       userInfo.Name,
					"avatar_url": userInfo.AvatarURL,
					"updated_at": now,
				})
		} else {
			return res.Error("Failed to link OAuth account: "+err.Error(), 500)
		}
	}

	// Update user avatar if not set
	var user models.User
	if err := core.DB().First(&user, userID).Error; err == nil {
		if user.Avatar == "" && userInfo.AvatarURL != "" {
			core.DB().Model(&user).Update("avatar", userInfo.AvatarURL)
		}
	}

	return res.Success(map[string]interface{}{
		"message": "OAuth account linked",
		"linked":  true,
	})
}
