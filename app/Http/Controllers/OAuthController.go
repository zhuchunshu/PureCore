package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	middleware "purecore/app/Http/Middleware"
	models "purecore/app/Models"
	"purecore/core"
	"purecore/core/oauth"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// OAuthController handles third-party OAuth login flows.
type OAuthController struct{}

// Providers lists all registered OAuth providers with their display info and enabled status.
func (c *OAuthController) Providers(req *core.Request, res *core.Response) error {
	providers := oauth.All()
	list := make([]map[string]interface{}, 0, len(providers))
	for _, p := range providers {
		enabled := core.IsOptionTrue(core.AdminOption(oauth.FullConfigKey(p, "enabled")))
		loginEnabled := core.IsOptionTrue(core.AdminOption(oauth.FullConfigKey(p, "login_enabled")))
		registerEnabled := core.IsOptionTrue(core.AdminOption(oauth.FullConfigKey(p, "register_enabled")))
		list = append(list, map[string]interface{}{
			"name":             p.Name(),
			"display_name":     p.DisplayName(),
			"enabled":          enabled,
			"login_enabled":    loginEnabled,
			"register_enabled": registerEnabled,
			"is_oauth2":        p.IsOAuth2(),
		})
	}
	return res.Success(list)
}

// Authorize generates the OAuth authorization URL for a provider and returns it.
// Accepts redirect via query string (GET) or JSON body (POST).
// Defaults to "/" if neither is provided.
func (c *OAuthController) Authorize(req *core.Request, res *core.Response) error {
	providerName := req.Ctx().Params("provider")
	provider := oauth.Get(providerName)
	if provider == nil {
		return res.Error("Unsupported OAuth provider", 400)
	}

	// Read redirect from query parameter first (used by frontend GET requests),
	// then fall back to JSON body for backward compatibility.
	redirect := req.Ctx().Query("redirect", "")
	if redirect == "" {
		type AuthRequest struct {
			Redirect string `json:"redirect"`
		}
		var body AuthRequest
		if err := req.Validate(&body); err == nil && body.Redirect != "" {
			redirect = body.Redirect
		}
	}
	if redirect == "" {
		redirect = "/"
	}

	// Default action is "login"
	action := "login"

	state, err := oauth.GenerateState(providerName, redirect, action)
	if err != nil {
		return res.Error("Failed to generate state", 500)
	}

	// For non-OAuth2 providers (e.g., Telegram), return widget config
	// instead of a redirect URL — the frontend renders the login widget.
	if !provider.IsOAuth2() {
		botUsername := ""
		if bp, ok := provider.(interface{ BotUsername() string }); ok {
			botUsername = bp.BotUsername()
		}
		if botUsername == "" {
			// Derive from bot_token: Telegram tokens are "123456:ABC..."
			botToken := oauth.GetProviderOption(provider.(oauth.Provider), "bot_token")
			botUsername = deriveBotUsernameFromToken(botToken)
		}
		redirectURL := oauth.GetProviderOption(provider.(oauth.Provider), "redirect_url")
		botID := deriveBotIDFromToken(oauth.GetProviderOption(provider.(oauth.Provider), "bot_token"))
		return res.Success(map[string]interface{}{
			"type":         "widget",
			"provider":     providerName,
			"state":        state,
			"bot_username": botUsername,
			"bot_id":       botID,
			"redirect_url": redirectURL,
		})
	}

	authURL := provider.GetAuthURL(state)
	return res.Success(map[string]interface{}{
		"url": authURL,
	})
}

// Callback handles the OAuth provider redirect after user authorization.
// This is the endpoint the provider redirects to after user consent.
// For OAuth2 providers, it exchanges the authorization code.
// For non-OAuth2 providers (e.g., Telegram), it verifies the signed callback data.
func (c *OAuthController) Callback(req *core.Request, res *core.Response) error {
	providerName := req.Ctx().Params("provider")
	provider := oauth.Get(providerName)
	if provider == nil {
		return res.Error("Unsupported OAuth provider", 400)
	}

	stateStr := req.Ctx().Query("state")
	code := req.Ctx().Query("code")

	if provider.IsOAuth2() {
		if code == "" || stateStr == "" {
			return res.Error("Missing code or state", 422)
		}
	} else {
		if stateStr == "" {
			return res.Error("Missing state", 422)
		}
	}

	// Parse state token (required for both flows)
	state, err := oauth.ParseState(stateStr)
	if err != nil {
		return res.Error("Invalid or expired state token", 422)
	}

	var userInfo *oauth.UserInfo

	if provider.IsOAuth2() {
		exchangeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		userInfo, err = provider.Exchange(exchangeCtx, code)
		if err != nil {
			// Log detailed error internally but return a generic message to the client
			return res.Error("OAuth authorization failed. Please try again.", 500)
		}
	} else {
		// Non-OAuth2 provider: collect query parameters (exclude state which is handled separately)
		params := make(map[string]string)
		req.Ctx().Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			k := string(key)
			if k != "state" { // state is validated separately above
				params[k] = string(value)
			}
		})
		userInfo, err = provider.HandleCallback(params)
		if err != nil {
			return res.Error("OAuth authorization failed. Please try again.", 500)
		}
	}

	// Check if this OAuth account is already linked to a user
	var existingOAuth models.OAuthAccount
	alreadyLinked := false
	var linkedUserID uint
	if err := core.DB().Where("provider = ? AND provider_id = ?", providerName, userInfo.ProviderID).First(&existingOAuth).Error; err == nil {
		if existingOAuth.UserID != 0 {
			alreadyLinked = true
			linkedUserID = existingOAuth.UserID
		}
	}

	// Base redirect URL: prefer the admin-configured site URL
	frontendBase := getSiteBaseURL(req)
	originalRedirect := state.Redirect
	if originalRedirect == "" {
		originalRedirect = "/"
	}

	// Case 1: Already linked → log the user in and redirect home
	if alreadyLinked {
		var user models.User
		if err := core.DB().First(&user, linkedUserID).Error; err != nil {
			// User was deleted — clean up the dangling OAuth record and continue to unlinked flow
			core.DB().Model(&existingOAuth).Update("user_id", 0)
			alreadyLinked = false
		} else {
			return c.loginWithRedirect(req, res, &user, providerName, frontendBase+"/oauth/"+providerName+"/callback?status=linked&redirect="+urlQueryEscape(originalRedirect))
		}
	}

	// Check if there's a logged-in user (optional).
	// The Callback route is public, so try: middleware locals → Auth header → cookie.
	currentUserID := getUserID(req.Ctx())
	if currentUserID == 0 {
		currentUserID = extractUserIDFromAuthHeader(req.Ctx())
	}
	if currentUserID == 0 {
		currentUserID = extractUserIDFromCookie(req.Ctx())
	}

	// Generate link token first (needed for both logged_in and unlinked paths)
	linkToken, err := oauth.GenerateLinkToken(providerName, userInfo)
	if err != nil {
		return res.Error("Failed to generate link token", 500)
	}

	// Case 2: Logged-in user → redirect with logged_in status so frontend asks whether to bind
	if currentUserID != 0 {
		var currentUser models.User
		redirectQuery := "status=logged_in&link_token=" + urlQueryEscape(linkToken) +
			"&provider=" + urlQueryEscape(providerName) +
			"&email=" + urlQueryEscape(userInfo.Email) +
			"&name=" + urlQueryEscape(userInfo.Name) +
			"&avatar_url=" + urlQueryEscape(userInfo.AvatarURL) +
			"&redirect=" + urlQueryEscape(originalRedirect)
		if err := core.DB().First(&currentUser, currentUserID).Error; err == nil {
			redirectQuery += "&current_user_name=" + urlQueryEscape(currentUser.Name) +
				"&current_user_email=" + urlQueryEscape(currentUser.Email)
		}
		redirectURL := frontendBase + "/oauth/" + providerName + "/callback?" + redirectQuery
		return req.Ctx().Redirect().To(redirectURL)
	}

	// Case 3: Not linked and not logged in → redirect with unlinked status
	redirectURL := frontendBase + "/oauth/" + providerName + "/callback?status=unlinked&link_token=" + urlQueryEscape(linkToken) +
		"&email=" + urlQueryEscape(userInfo.Email) +
		"&name=" + urlQueryEscape(userInfo.Name) +
		"&avatar_url=" + urlQueryEscape(userInfo.AvatarURL) +
		"&redirect=" + urlQueryEscape(originalRedirect)
	return req.Ctx().Redirect().To(redirectURL)
}

// Exchange handles the OAuth code exchange initiated by the frontend callback page.
// The frontend receives code+state from the OAuth provider redirect and calls this
// endpoint to complete the OAuth flow. Returns JSON with appropriate status:
//   - "linked": OAuth account already linked → tokens returned, user logged in
//   - "logged_in": User is already logged in → link_token + user info returned, frontend asks whether to bind
//   - "unlinked": New OAuth account → link_token + user info returned for registration/binding
func (c *OAuthController) Exchange(req *core.Request, res *core.Response) error {
	providerName := req.Ctx().Params("provider")
	provider := oauth.Get(providerName)
	if provider == nil {
		return res.Error("Unsupported OAuth provider", 400)
	}

	var stateToken string
	if provider.IsOAuth2() {
		type ExchangeOAuth2Request struct {
			Code  string `json:"code" validate:"required"`
			State string `json:"state" validate:"required"`
		}
		var body ExchangeOAuth2Request
		if err := req.Validate(&body); err != nil {
			return res.Error("Missing code or state", 422)
		}
		stateToken = body.State
	} else {
		type ExchangeNonOAuth2Request struct {
			State string `json:"state" validate:"required"`
		}
		var body ExchangeNonOAuth2Request
		if err := req.Validate(&body); err != nil {
			return res.Error("Missing state for non-OAuth2 provider", 422)
		}
		stateToken = body.State
	}

	state, err := oauth.ParseState(stateToken)
	if err != nil {
		return res.Error("Invalid or expired state token", 422)
	}

	var userInfo *oauth.UserInfo

	if provider.IsOAuth2() {
		type ExchangeOAuth2Request struct {
			Code  string `json:"code" validate:"required"`
			State string `json:"state" validate:"required"`
		}
		var body ExchangeOAuth2Request
		if err := req.Validate(&body); err != nil {
			return res.Error("Missing code or state", 422)
		}
		exchangeCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		userInfo, err = provider.Exchange(exchangeCtx, body.Code)
		if err != nil {
			return res.Error("Failed to exchange OAuth code: "+err.Error(), 500)
		}
	} else {
		// Non-OAuth2 provider (e.g., Telegram): receive callback params in request body
		type NonOAuth2Params struct {
			State string            `json:"state" validate:"required"`
			Data  map[string]string `json:"data"`
		}
		var nonOAuth2Body NonOAuth2Params
		if err := req.Validate(&nonOAuth2Body); err != nil {
			return res.Error("Missing state or callback data for non-OAuth2 provider", 422)
		}
		state, err = oauth.ParseState(nonOAuth2Body.State)
		if err != nil {
			return res.Error("Invalid or expired state token", 422)
		}
		if nonOAuth2Body.Data == nil || len(nonOAuth2Body.Data) == 0 {
			// Try to read raw body as map
			var rawBody map[string]interface{}
			if err := json.Unmarshal([]byte(req.Ctx().Body()), &rawBody); err == nil {
				nonOAuth2Body.Data = make(map[string]string)
				for k, v := range rawBody {
					if k == "state" {
						continue
					}
					if s, ok := v.(string); ok {
						nonOAuth2Body.Data[k] = s
					}
				}
			}
		}
		userInfo, err = provider.HandleCallback(nonOAuth2Body.Data)
		if err != nil {
			return res.Error("OAuth authorization failed: "+err.Error(), 500)
		}
	}

	originalRedirect := state.Redirect
	if originalRedirect == "" {
		originalRedirect = "/"
	}

	// Check if already linked (ignore dangling records with UserID == 0)
	var existingOAuth models.OAuthAccount
	if err := core.DB().Where("provider = ? AND provider_id = ?", providerName, userInfo.ProviderID).First(&existingOAuth).Error; err == nil {
		if existingOAuth.UserID != 0 {
			var user models.User
			if err := core.DB().First(&user, existingOAuth.UserID).Error; err != nil {
				// Dangling reference — clean up and continue to unlinked flow
				core.DB().Model(&existingOAuth).Update("user_id", 0)
			} else {
				// Login and return tokens
				return c.loginUser(req, res, &user, providerName, originalRedirect)
			}
		}
	}

	// Generate link token first (needed for both logged_in and unlinked paths)
	linkToken, err := oauth.GenerateLinkToken(providerName, userInfo)
	if err != nil {
		return res.Error("Failed to generate link token", 500)
	}

	// Check if there's a logged-in user.
	// The Exchange route is public (no auth middleware), so Locals("user") may
	// be empty even when a valid Bearer token is present. Try: locals → auth header → cookie.
	currentUserID := getUserID(req.Ctx())
	if currentUserID == 0 {
		currentUserID = extractUserIDFromAuthHeader(req.Ctx())
	}
	if currentUserID == 0 {
		currentUserID = extractUserIDFromCookie(req.Ctx())
	}
	if currentUserID != 0 {
		// User is already logged in — return logged_in status so the frontend
		// can ask the user whether to bind this OAuth account.
		var currentUser models.User
		response := map[string]interface{}{
			"status":     "logged_in",
			"link_token": linkToken,
			"provider":   providerName,
			"email":      userInfo.Email,
			"name":       userInfo.Name,
			"avatar_url": userInfo.AvatarURL,
			"redirect":   originalRedirect,
		}
		if err := core.DB().First(&currentUser, currentUserID).Error; err == nil {
			response["current_user"] = map[string]interface{}{
				"name":  currentUser.Name,
				"email": currentUser.Email,
			}
		}
		return res.Success(response)
	}

	// Not linked and not logged in → return unlinked status
	return res.Success(map[string]interface{}{
		"status":     "unlinked",
		"link_token": linkToken,
		"provider":   providerName,
		"email":      userInfo.Email,
		"name":       userInfo.Name,
		"avatar_url": userInfo.AvatarURL,
		"redirect":   originalRedirect,
	})
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
	return c.loginUser(req, res, &user, oAuthInfo.Provider, "/")
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

	// Check if this provider ID is already linked (including soft-deleted rows).
	// The DB unique index is on provider_id, so we must query by provider_id only.
	var existing models.OAuthAccount
	now := time.Now()
	if err := core.DB().Unscoped().Where("provider_id = ?", oAuthInfo.ProviderID).First(&existing).Error; err == nil {
		// Soft-deleted historical links should be reclaimable after fresh OAuth auth.
		if !existing.DeletedAt.Valid && existing.UserID != userID {
			return res.Error("This OAuth account is already linked to another user", 409)
		}

		updates := map[string]interface{}{
			"user_id":       userID,
			"provider":      oAuthInfo.Provider,
			"email":         oAuthInfo.Email,
			"name":          oAuthInfo.Name,
			"avatar_url":    oAuthInfo.AvatarURL,
			"access_token":  oAuthInfo.AccessToken,
			"token_expiry":  oAuthInfo.TokenExpiry,
			"raw_data":      "",
			"refresh_token": "",
			"updated_at":    now,
			"deleted_at":    nil, // restore soft-deleted links
		}
		if oAuthInfo.RawData != nil {
			rawBytes, _ := json.Marshal(oAuthInfo.RawData)
			updates["raw_data"] = string(rawBytes)
		}

		if err := core.DB().Unscoped().Model(&existing).Updates(updates).Error; err != nil {
			return res.Error("Failed to link OAuth account: "+err.Error(), 500)
		}

		// Update user's avatar if not set
		var user models.User
		if err := core.DB().First(&user, userID).Error; err == nil {
			if user.Avatar == "" && oAuthInfo.AvatarURL != "" {
				core.DB().Model(&user).Update("avatar", oAuthInfo.AvatarURL)
			}
		}

		return res.Success("Already linked")
	}

	// Create the link
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
		// Handle race conditions gracefully: another request may have linked it meanwhile.
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			var raceExisting models.OAuthAccount
			if qErr := core.DB().Unscoped().Where("provider_id = ?", oAuthInfo.ProviderID).First(&raceExisting).Error; qErr == nil {
				if !raceExisting.DeletedAt.Valid && raceExisting.UserID != userID {
					return res.Error("This OAuth account is already linked to another user", 409)
				}

				updates := map[string]interface{}{
					"user_id":       userID,
					"provider":      oAuthInfo.Provider,
					"email":         oAuthInfo.Email,
					"name":          oAuthInfo.Name,
					"avatar_url":    oAuthInfo.AvatarURL,
					"access_token":  oAuthInfo.AccessToken,
					"token_expiry":  oAuthInfo.TokenExpiry,
					"raw_data":      "",
					"refresh_token": "",
					"updated_at":    now,
					"deleted_at":    nil,
				}
				if oAuthInfo.RawData != nil {
					rawBytes, _ := json.Marshal(oAuthInfo.RawData)
					updates["raw_data"] = string(rawBytes)
				}
				if uErr := core.DB().Unscoped().Model(&raceExisting).Updates(updates).Error; uErr != nil {
					return res.Error("Failed to link OAuth account: "+uErr.Error(), 500)
				}
			} else {
				return res.Error("Failed to link OAuth account: "+err.Error(), 500)
			}
		} else {
			return res.Error("Failed to link OAuth account: "+err.Error(), 500)
		}
	}

	// Update user's avatar if not set
	var user models.User
	if err := core.DB().First(&user, userID).Error; err == nil {
		if user.Avatar == "" && oAuthInfo.AvatarURL != "" {
			core.DB().Model(&user).Update("avatar", oAuthInfo.AvatarURL)
		}
	}

	return res.Success(map[string]interface{}{
		"message": "OAuth account linked successfully",
	})
}

// Accounts returns all OAuth accounts linked to the current user,
// along with the login provider of the current session (if any).
func (c *OAuthController) Accounts(req *core.Request, res *core.Response) error {
	userID := getUserID(req.Ctx())
	if userID == 0 {
		return res.Unauthorized()
	}

	var accounts []models.OAuthAccount
	core.DB().Where("user_id = ?", userID).Find(&accounts)

	// Look up the current session to determine which OAuth provider was used for this login
	var currentSession models.UserSession
	currentLoginProvider := ""
	if err := core.DB().Where("user_id = ? AND is_current = ?", userID, true).First(&currentSession).Error; err == nil {
		currentLoginProvider = currentSession.LoginProvider
	}

	return res.Success(map[string]interface{}{
		"accounts":              accounts,
		"current_login_provider": currentLoginProvider,
	})
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
// Each provider entry includes its current setting values, dynamic ConfigFields metadata,
// documentation/application URLs, and the recommended callback URL based on the site URL.
func (c *OAuthController) AdminGetSettings(req *core.Request, res *core.Response) error {
	providers := oauth.All()
	settings := make([]map[string]interface{}, 0, len(providers))

	siteURL := getSiteBaseURL(req)

	for _, p := range providers {
		prefix := p.ConfigKeyPrefix()
		configFields := p.ConfigFields()

		// Build current values map for all configured fields
		values := make(map[string]string)
		for _, field := range configFields {
			fullKey := prefix + "_" + field.Key
			values[field.Key] = core.AdminOption(fullKey, "")
		}
		// For standard OAuth2 providers, auto-fill the redirect_url default if empty
		if p.IsOAuth2() {
			recommendedRedirect := siteURL + oauth.RedirectURLPath(p.Name())
			if values["redirect_url"] == "" {
				values["redirect_url"] = recommendedRedirect
			}
			// Add a hint about the recommended default
			values["_recommended_redirect_url"] = recommendedRedirect
		} else {
			// Non-OAuth2 providers: suggest a frontend callback URL
			recommendedRedirect := siteURL + "/oauth/" + p.Name() + "/callback"
			if values["redirect_url"] == "" {
				values["redirect_url"] = recommendedRedirect
			}
			values["_recommended_redirect_url"] = recommendedRedirect
		}

		providerSettings := map[string]interface{}{
			"name":              p.Name(),
			"display_name":      p.DisplayName(),
			"config_key_prefix": prefix,
			"is_oauth2":         p.IsOAuth2(),
			"doc_url":           p.GetDocURL(),
			"apply_url":         p.GetApplyURL(),
			"config_fields":     configFields,
			"values":            values,
		}
		settings = append(settings, providerSettings)
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

// loginUser generates tokens and creates a session for a user, returning JSON
// with status "linked" so the frontend can process the OAuth exchange result correctly.
// providerName records which OAuth provider was used for this login session.
func (c *OAuthController) loginUser(req *core.Request, res *core.Response, user *models.User, providerName string, redirect string) error {
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
	CreateSession(req.Ctx(), user.ID, providerName)

	return res.Success(map[string]interface{}{
		"status":        "linked",
		"token":         accessToken,
		"refresh_token": refreshToken,
		"name":          user.Name,
		"email":         user.Email,
		"redirect":      redirect,
	})
}

// loginWithRedirect logs the user in and sets tokens as readable cookies,
// then redirects to the given URL. Cookies are NOT HttpOnly so the frontend
// JS can read them on the callback page. providerName records which OAuth
// provider was used for this login session.
func (c *OAuthController) loginWithRedirect(req *core.Request, res *core.Response, user *models.User, providerName string, redirectTo string) error {
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
	CreateSession(req.Ctx(), user.ID, providerName)

	// Set tokens as non-HttpOnly cookies so the frontend JS can read them
	// on the OAuth callback page to complete the login flow.
	isProduction := core.GetConfig().IsProduction()
	req.Ctx().Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Secure:   isProduction,
		HTTPOnly: false,
		SameSite: "Lax",
	})
	req.Ctx().Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Secure:   isProduction,
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

// getSiteBaseURL returns the configured base URL of the site (no trailing slash).
// Priority: 1) Admin-configured "app_url" option, 2) FRONTEND_URL env var,
// 3) Auto-detection from the current request.
func getSiteBaseURL(req *core.Request) string {
	if appURL := strings.TrimRight(core.AdminOption("app_url", ""), "/"); appURL != "" {
		return appURL
	}
	// Check for FRONTEND_URL environment variable (for dev environments)
	if frontendURL := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/"); frontendURL != "" {
		return frontendURL
	}
	// Fallback: auto-detect from request
	scheme := req.Ctx().Protocol()
	host := req.Ctx().Hostname()
	return scheme + "://" + host
}

func urlQueryEscape(s string) string {
	return url.QueryEscape(s)
}

// deriveBotUsernameFromToken extracts the bot username from a Telegram bot token.
// Telegram bot tokens follow the format: "NUMBER:ALPHANUMERIC_STRING"
// The bot username cannot be derived from the token alone, so this returns ""
// as a fallback — the frontend should prompt for the username if needed.
func deriveBotUsernameFromToken(token string) string {
	// Bot username cannot be extracted from the token.
	// The admin must configure it, or we can try common patterns.
	// For now return empty; the frontend can use the widget without it
	// as the Telegram widget script can auto-detect from data-* attributes.
	return ""
}

// deriveBotIDFromToken extracts the numeric bot ID prefix from a Telegram bot token.
// Telegram bot tokens follow the format: "<BOT_ID>:<SECRET>".
func deriveBotIDFromToken(token string) string {
	parts := strings.SplitN(strings.TrimSpace(token), ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[0])
}

// extractUserIDFromAuthHeader tries to parse a user JWT from the Authorization
// header. This is used on public OAuth routes where the auth middleware is not
// applied, so a logged-in user's token wouldn't be populated in Locals("user").
// Returns the user ID if a valid token is found, or 0 otherwise.
func extractUserIDFromAuthHeader(c fiber.Ctx) uint {
	auth := c.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
		return 0
	}
	return parseUserIDFromToken(strings.TrimPrefix(auth, "Bearer "))
}

// extractUserIDFromCookie tries to parse a user JWT from the access_token cookie.
// This is used as a fallback on public OAuth routes that receive browser redirects
// (no Authorization header). The cookie may have been set by loginWithRedirect().
// Returns the user ID if a valid token is found, or 0 otherwise.
func extractUserIDFromCookie(c fiber.Ctx) uint {
	tokenStr := c.Cookies("access_token", "")
	if tokenStr == "" {
		return 0
	}
	return parseUserIDFromToken(tokenStr)
}

// parseUserIDFromToken validates a JWT token string and extracts the user_id claim.
// Returns the user ID if the token is valid, or 0 otherwise.
func parseUserIDFromToken(tokenStr string) uint {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims,
		func(t *jwt.Token) (interface{}, error) {
			return []byte(core.GetConfig().String("JWT_SECRET")), nil
		},
	)
	if err != nil || !token.Valid {
		return 0
	}
	if id, ok := claims["user_id"]; ok {
		switch v := id.(type) {
		case float64:
			if v > 0 && v <= math.MaxUint32 {
				return uint(v)
			}
		case string:
			if parsed, err := strconv.ParseUint(v, 10, 32); err == nil {
				return uint(parsed)
			}
		}
	}
	return 0
}
