package oauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AppleProvider implements Sign in with Apple (OAuth2 with JWT client secret).
// Docs: https://developer.apple.com/documentation/sign_in_with_apple
// App creation: https://developer.apple.com/account/resources/identifiers/list/serviceId
type AppleProvider struct {
	BaseProvider
}

func init() {
	Register(&AppleProvider{})
}

func (p *AppleProvider) Name() string            { return "apple" }
func (p *AppleProvider) DisplayName() string     { return "Apple" }
func (p *AppleProvider) ConfigKeyPrefix() string { return "oauth_apple" }

func (p *AppleProvider) GetDocURL() string {
	return "https://developer.apple.com/documentation/sign_in_with_apple"
}

func (p *AppleProvider) GetApplyURL() string {
	return "https://developer.apple.com/account/resources/identifiers/list/serviceId"
}

// ConfigFields returns extended fields for Apple (adds private_key, team_id, key_id).
func (p *AppleProvider) ConfigFields() []ConfigField {
	fields := BaseOAuth2ConfigFields()
	fields = append(fields,
		ConfigField{Key: "team_id", Label: "admin.oauth_apple_team_id", Type: "text", Required: true},
		ConfigField{Key: "key_id", Label: "admin.oauth_apple_key_id", Type: "text", Required: true},
		ConfigField{Key: "private_key", Label: "admin.oauth_apple_private_key", Type: "password", Required: true, Help: "admin.oauth_apple_private_key_help"},
	)
	return fields
}

// GetAuthURL builds the Apple Sign In authorize URL.
func (p *AppleProvider) GetAuthURL(state string) string {
	clientID := GetProviderOption(p, "client_id")
	redirectURL := GetProviderOption(p, "redirect_url")
	u, _ := url.Parse("https://appleid.apple.com/auth/authorize")
	q := u.Query()
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "name email")
	q.Set("state", state)
	q.Set("response_mode", "form_post")
	u.RawQuery = q.Encode()
	return u.String()
}

// Exchange exchanges the authorization code for an access token and ID token, then extracts user info.
func (p *AppleProvider) Exchange(ctx context.Context, code string) (*UserInfo, error) {
	clientID := GetProviderOption(p, "client_id")
	teamID := GetProviderOption(p, "team_id")
	keyID := GetProviderOption(p, "key_id")
	privateKey := GetProviderOption(p, "private_key")
	redirectURL := GetProviderOption(p, "redirect_url")

	// Generate client secret JWT
	clientSecret, err := p.generateClientSecret(clientID, teamID, keyID, privateKey)
	if err != nil {
		return nil, fmt.Errorf("apple: failed to generate client secret: %w", err)
	}

	// Exchange code for token
	tokenResp, err := p.exchangeCode(ctx, clientID, clientSecret, redirectURL, code)
	if err != nil {
		return nil, fmt.Errorf("apple: failed to exchange code: %w", err)
	}

	// Extract user info from ID token
	user, err := p.parseIDToken(tokenResp.IDToken)
	if err != nil {
		return nil, fmt.Errorf("apple: failed to parse ID token: %w", err)
	}

	return &UserInfo{
		ProviderID: user.Sub,
		Email:      user.Email,
		Name:       user.Name,
		AvatarURL:  "",
		Raw:        user,
	}, nil
}

// HandleCallback is not used for Apple (standard OAuth2).
func (p *AppleProvider) HandleCallback(params map[string]string) (*UserInfo, error) {
	return nil, fmt.Errorf("Apple uses OAuth2 code exchange; use Exchange instead")
}

// ---------- Apple config fields and helpers ----------

type appleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	Error        string `json:"error"`
}

type appleIDTokenClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Name  string `json:"name"`
	Sub   string `json:"sub"` // Unique user ID
}

func (p *AppleProvider) generateClientSecret(clientID, teamID, keyID, privateKey string) (string, error) {
	// Parse the private key
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the private key")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}
	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("private key is not an ECDSA key")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss": teamID,
		"iat": now.Unix(),
		"exp": now.Add(24 * time.Hour).Unix(),
		"aud": "https://appleid.apple.com",
		"sub": clientID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = keyID
	return token.SignedString(ecdsaKey)
}

func (p *AppleProvider) exchangeCode(ctx context.Context, clientID, clientSecret, redirectURL, code string) (*appleTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURL)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://appleid.apple.com/auth/token", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tr appleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}
	if tr.Error != "" {
		return nil, fmt.Errorf("apple token error: %s", tr.Error)
	}
	return &tr, nil
}

func (p *AppleProvider) parseIDToken(idToken string) (*appleIDTokenClaims, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(idToken, &appleIDTokenClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse id_token: %w", err)
	}
	claims, ok := token.Claims.(*appleIDTokenClaims)
	if !ok {
		return nil, fmt.Errorf("unexpected claims type")
	}
	return claims, nil
}
