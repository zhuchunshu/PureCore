package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	middleware "purecore/app/Http/Middleware"
	models "purecore/app/Models"
	"purecore/core"

	"github.com/gofiber/fiber/v3"
	"github.com/markbates/goth"
)

// OAuthController handles OAuth login flows for all enabled providers.
type OAuthController struct{}

// oauthStateTTL is how long the state cookie is valid
const oauthStateTTL = 15 * time.Minute

// Redirect initiates the OAuth flow by building the provider's authorization URL
// and redirecting the user's browser. It sets a state cookie for CSRF protection.
func (oc *OAuthController) Redirect(req *core.Request, res *core.Response) error {
	provider := req.Ctx().Params("provider")

	prov, err := goth.GetProvider(provider)
	if err != nil {
		return res.Error("Unsupported OAuth provider: "+provider, 400)
	}

	// Verify login is allowed for this provider
	if !core.OAuthAllowLogin(provider) {
		return res.Error("OAuth login is not enabled for "+provider, 403)
	}

	// Generate random state for CSRF protection
	state, err := generateState()
	if err != nil {
		return res.Error("Failed to generate state", 500)
	}

	// Store state in a cookie
	req.Ctx().Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(oauthStateTTL),
		HTTPOnly: true,
		Secure:   isHTTPS(req.Ctx()),
		SameSite: "Lax",
	})

	// Build the authorization URL via Goth
	sess, err := prov.BeginAuth(state)
	if err != nil {
		return res.Error("Failed to begin auth: "+err.Error(), 500)
	}

	authURL, err := sess.GetAuthURL()
	if err != nil {
		return res.Error("Failed to get auth URL: "+err.Error(), 500)
	}

	return req.Ctx().Redirect().To(authURL)
}

// Callback handles the OAuth callback after the user authorizes the app.
func (oc *OAuthController) Callback(req *core.Request, res *core.Response) error {
	provider := req.Ctx().Params("provider")
	code := req.Ctx().Query("code")
	state := req.Ctx().Query("state")

	prov, err := goth.GetProvider(provider)
	if err != nil {
		return res.Error("Unsupported OAuth provider: "+provider, 400)
	}

	// Verify login is still allowed (could have been disabled mid-flow)
	if !core.OAuthAllowLogin(provider) {
		return res.Error("OAuth login is not enabled for "+provider, 403)
	}

	// For Telegram, the authentication data is sent as query parameters without a code.
	// Standard OAuth2 providers use code + state.
	if provider == "telegram" {
		return oc.handleTelegramCallback(req, res, prov)
	}

	// ----- Standard OAuth2 flow (GitHub, Google, Discord, Apple) -----

	// Verify state to prevent CSRF
	cookieState := req.Ctx().Cookies("oauth_state")
	if cookieState == "" || state == "" || state != cookieState {
		return res.Error("Invalid OAuth state parameter", 400)
	}

	// Clear the state cookie
	req.Ctx().Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
	})

	// Complete the OAuth flow via Goth
	sess, err := prov.BeginAuth(state)
	if err != nil {
		return res.Error("Failed to begin auth session: "+err.Error(), 500)
	}

	// Exchange code for token and fetch user info
	_, err = sess.Authorize(prov, url.Values{"code": {code}, "state": {state}})
	if err != nil {
		return res.Error("Failed to authorize: "+err.Error(), 500)
	}

	return oc.completeLogin(req.Ctx(), prov, sess, provider)
}

// handleTelegramCallback processes the Telegram OAuth callback which uses query parameters
// instead of the OAuth2 authorization code flow.
func (oc *OAuthController) handleTelegramCallback(req *core.Request, res *core.Response, prov goth.Provider) error {
	// Telegram doesn't use a code; the auth data is in the query string
	// We create a session and pass the query params directly to Authorize.
	sess, err := prov.BeginAuth("") // State not used for Telegram
	if err != nil {
		return res.Error("Failed to create telegram session: "+err.Error(), 500)
	}

	// Convert all query parameters into goth.Params (which is url.Values)
	queryMap := req.Ctx().Queries()
	params := url.Values{}
	for k, v := range queryMap {
		params.Add(k, v)
	}
	_, err = sess.Authorize(prov, params)
	if err != nil {
		return res.Error("Failed to authorize Telegram: "+err.Error(), 500)
	}

	return oc.completeLogin(req.Ctx(), prov, sess, "telegram")
}

// completeLogin fetches the Goth user, finds or creates a local user, and issues JWT tokens.
func (oc *OAuthController) completeLogin(c fiber.Ctx, prov goth.Provider, sess goth.Session, provider string) error {
	gothUser, err := prov.FetchUser(sess)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to fetch user: " + err.Error()})
	}

	// Determine email and name from Goth user
	email := gothUser.Email
	if email == "" {
		email = fmt.Sprintf("%s@%s.user", gothUser.NickName, provider)
	}
	name := gothUser.Name
	if name == "" {
		name = gothUser.NickName
		if name == "" {
			name = gothUser.FirstName + " " + gothUser.LastName
		}
	}
	if name == "" || name == " " {
		name = email
	}

	// Find or create user
	user, isNew, err := oc.findOrCreateUser(provider, gothUser.UserID, email, name, gothUser.AvatarURL)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to find or create user: " + err.Error()})
	}

	// If trying to register a new account but registration is disabled, reject
	if isNew && !core.OAuthAllowRegister(provider) {
		return c.Status(403).JSON(fiber.Map{"code": 403, "message": "OAuth registration is disabled for " + provider})
	}

	// Save/update the OAuth provider binding
	rawJSON, _ := json.Marshal(gothUser.RawData)
	oc.saveProviderBinding(user.ID, provider, gothUser.UserID, email,
		gothUser.NickName, gothUser.AccessToken, "", "", gothUser.AvatarURL, string(rawJSON))

	// Generate JWT tokens
	accessTokenJWT, err := middleware.GenerateUserToken(user.ID, user.Name)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to generate token: " + err.Error()})
	}

	refreshToken, err := middleware.GenerateRefreshToken()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "Failed to generate refresh token: " + err.Error()})
	}

	// Save refresh token and update last login
	now := time.Now()
	core.DB().Model(user).Updates(map[string]interface{}{
		"refresh_token": refreshToken,
		"last_login_at": now,
	})

	// When creating a new user via OAuth, also set the email_verified_at timestamp
	if isNew {
		core.DB().Model(user).Update("email_verified_at", now)
	}

	// Create session
	CreateSession(c, user.ID)

	// Build frontend redirect URL with tokens
	frontendURL := core.GetConfig().String("APP_URL", "http://localhost:3000")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}
	redirectURL := fmt.Sprintf("%s/dashboard?token=%s&refresh_token=%s&name=%s&email=%s",
		frontendURL,
		url.QueryEscape(accessTokenJWT),
		url.QueryEscape(refreshToken),
		url.QueryEscape(user.Name),
		url.QueryEscape(user.Email),
	)

	return c.Redirect().To(redirectURL)
}

// EnabledProviders returns which OAuth providers are configured and active
func (oc *OAuthController) EnabledProviders(req *core.Request, res *core.Response) error {
	providers := core.EnabledOAuthProviders()
	return res.Success(providers)
}

// findOrCreateUser finds an existing user by their OAuth binding or email,
// or creates a new user
func (oc *OAuthController) findOrCreateUser(provider, providerUserID, email, name, avatar string) (*models.User, bool, error) {
	// Check if this OAuth binding already exists
	var binding models.UserOauthProvider
	err := core.DB().Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&binding).Error
	if err == nil {
		// Found existing binding — look up the user
		var user models.User
		if err := core.DB().First(&user, binding.UserID).Error; err == nil {
			return &user, false, nil
		}
	}

	// Check if a user with this email already exists
	var existingUser models.User
	if err := core.DB().Where("email = ?", email).First(&existingUser).Error; err == nil {
		return &existingUser, false, nil
	}

	// Create a new user
	newUser := models.User{
		Name:  name,
		Email: email,
	}
	// OAuth users have no password — set a random one that cannot be used for login
	if err := newUser.SetPassword(generateRandomPassword(32)); err != nil {
		return nil, false, err
	}

	if avatar != "" {
		newUser.Avatar = avatar
	}

	if err := core.DB().Create(&newUser).Error; err != nil {
		return nil, false, err
	}

	return &newUser, true, nil
}

// saveProviderBinding creates or updates the OAuth provider binding record
func (oc *OAuthController) saveProviderBinding(userID uint, provider, providerUserID, email, username, accessToken, refreshToken, tokenExpiresAt, avatarURL, rawData string) {
	var binding models.UserOauthProvider
	err := core.DB().Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&binding).Error
	if err != nil {
		// Create new binding
		binding = models.UserOauthProvider{
			UserID:           userID,
			Provider:         provider,
			ProviderUserID:   providerUserID,
			ProviderEmail:    email,
			ProviderUsername: username,
			AccessToken:      accessToken,
			RefreshToken:     refreshToken,
			TokenExpiresAt:   tokenExpiresAt,
			AvatarURL:        avatarURL,
			RawData:          rawData,
		}
		core.DB().Create(&binding)
	} else {
		// Update existing binding
		core.DB().Model(&binding).Updates(map[string]interface{}{
			"user_id":           userID,
			"provider_email":    email,
			"provider_username": username,
			"access_token":      accessToken,
			"refresh_token":     refreshToken,
			"token_expires_at":  tokenExpiresAt,
			"avatar_url":        avatarURL,
			"raw_data":          rawData,
		})
	}
}

// -------- Helpers --------

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func generateRandomPassword(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func isHTTPS(c fiber.Ctx) bool {
	return c.Protocol() == "https" || c.Get("X-Forwarded-Proto") == "https"
}
