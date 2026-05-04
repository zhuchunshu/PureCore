package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// StateClaims represents the claims stored in the OAuth state token.
type StateClaims struct {
	jwt.RegisteredClaims
	Provider string `json:"provider"`
	Redirect string `json:"redirect"` // Frontend URL to redirect back to after callback
	CSRF     string `json:"csrf"`     // Random CSRF token
	Action   string `json:"action"`   // "login" or "register"
}

// GenerateState creates a signed JWT state token that is passed to the OAuth provider
// and verified on callback. It contains the provider name, a CSRF nonce, and the
// intended frontend redirect URL.
func GenerateState(provider, redirect, action string) (string, error) {
	csrf := make([]byte, 16)
	if _, err := rand.Read(csrf); err != nil {
		return "", err
	}

	claims := StateClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Provider: provider,
		Redirect: redirect,
		CSRF:     hex.EncodeToString(csrf),
		Action:   action,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret()))
}

// ParseState parses and validates a state token returned from an OAuth callback.
// Returns the claims if valid, or an error if expired, invalid signature, etc.
func ParseState(state string) (*StateClaims, error) {
	claims := &StateClaims{}
	token, err := jwt.ParseWithClaims(state, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret()), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// jwtSecret returns the JWT secret from environment variables.
// This is duplicated here to keep the oauth package self-contained,
// but should match the value used by the middleware package.
func jwtSecret() string {
	if defaultSecret == "" {
		panic("oauth.SetSecret() must be called during application boot with a non-empty secret")
	}
	return defaultSecret
}

var defaultSecret string

// SetSecret allows the application to inject the JWT secret for OAuth state tokens.
// Call this once during boot, e.g. from cmd/serve.go.
func SetSecret(secret string) {
	defaultSecret = secret
}

// ---------- Link Token (short-lived token carrying OAuth user info) ----------

// LinkTokenClaims stores the OAuth provider user info inside a signed JWT.
// This token is returned by the callback endpoint when the OAuth identity
// is not yet linked to any PureCore user. The frontend passes it back to
// the Register or Bind endpoints to complete the account linking.
type LinkTokenClaims struct {
	jwt.RegisteredClaims
	Provider    string `json:"provider"`
	ProviderID  string `json:"provider_id"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	AccessToken string `json:"access_token"`
}

// LinkTokenData is the deserialized representation returned by ParseLinkToken.
// Does not expose internal claims such as expiry.
type LinkTokenData struct {
	Provider    string     `json:"provider"`
	ProviderID  string     `json:"provider_id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	AvatarURL   string     `json:"avatar_url"`
	AccessToken string     `json:"access_token"`
	RawData     any        `json:"raw_data,omitempty"`
	TokenExpiry *time.Time `json:"token_expiry,omitempty"`
}

// GenerateLinkToken creates a short-lived (5 minute) JWT that securely carries
// the OAuth user profile from the callback endpoint to the register/bind flow.
func GenerateLinkToken(providerName string, info *UserInfo) (string, error) {
	claims := LinkTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Provider:    providerName,
		ProviderID:  info.ProviderID,
		Email:       info.Email,
		Name:        info.Name,
		AvatarURL:   info.AvatarURL,
		AccessToken: "", // Populated later if provider returns one
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret()))
}

// ParseLinkToken validates and decodes a link token.
func ParseLinkToken(tokenStr string) (*LinkTokenData, error) {
	claims := &LinkTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret()), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return &LinkTokenData{
		Provider:    claims.Provider,
		ProviderID:  claims.ProviderID,
		Email:       claims.Email,
		Name:        claims.Name,
		AvatarURL:   claims.AvatarURL,
		AccessToken: claims.AccessToken,
	}, nil
}

// ---------- State Data helpers ----------

// StateData is a serializable representation of the state claims, used to
// communicate the OAuth intent to the frontend after callback.
type StateData struct {
	Provider string `json:"provider"`
	Redirect string `json:"redirect"`
	Action   string `json:"action"`
}

// ToStateData converts claims to a frontend-friendly struct.
func (c *StateClaims) ToStateData() StateData {
	return StateData{
		Provider: c.Provider,
		Redirect: c.Redirect,
		Action:   c.Action,
	}
}

// MarshalStateData serializes state data to JSON bytes for embedding in the callback response.
func MarshalStateData(data StateData) ([]byte, error) {
	return json.Marshal(data)
}
