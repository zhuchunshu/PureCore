package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/markbates/goth"
	"golang.org/x/oauth2"
)

// TelegramProvider implements the goth.Provider interface for Telegram OAuth.
// Telegram uses a widget-based authentication flow where the user is redirected
// to the Telegram login widget, and the authorization data is sent back as
// query parameters to the callback URL.
type TelegramProvider struct {
	name        string
	botToken    string
	callbackURL string
}

// NewTelegramProvider creates a new Telegram OAuth provider.
func NewTelegramProvider(botToken, callbackURL string) *TelegramProvider {
	return &TelegramProvider{
		name:        "telegram",
		botToken:    botToken,
		callbackURL: callbackURL,
	}
}

// Name returns the provider name.
func (p *TelegramProvider) Name() string {
	return p.name
}

// SetName sets the provider name.
func (p *TelegramProvider) SetName(name string) {
	p.name = name
}

// Debug is a no-op.
func (p *TelegramProvider) Debug(debug bool) {}

// BeginAuth builds the Telegram login widget URL.
func (p *TelegramProvider) BeginAuth(state string) (goth.Session, error) {
	authURL := fmt.Sprintf("https://oauth.telegram.org/auth?bot_id=%s&origin=%s&embed=1&request_access=write",
		p.botToken,
		url.QueryEscape(p.callbackURL),
	)

	session := &TelegramSession{
		AuthURL: authURL,
	}
	return session, nil
}

// FetchUser validates the Telegram authorization data and returns a goth.User.
func (p *TelegramProvider) FetchUser(session goth.Session) (goth.User, error) {
	sess, ok := session.(*TelegramSession)
	if !ok {
		return goth.User{}, fmt.Errorf("invalid session type for telegram provider")
	}

	data := sess.AuthData
	if data == nil {
		return goth.User{}, fmt.Errorf("no auth data available")
	}

	if err := p.verifyAuthData(data); err != nil {
		return goth.User{}, fmt.Errorf("telegram auth verification failed: %w", err)
	}

	userID := fmt.Sprintf("%v", data["id"])
	firstName, _ := data["first_name"].(string)
	lastName, _ := data["last_name"].(string)
	username, _ := data["username"].(string)
	photoURL, _ := data["photo_url"].(string)

	name := firstName
	if lastName != "" {
		name = firstName + " " + lastName
	}
	if name == "" {
		name = username
	}
	if name == "" {
		name = userID
	}

	email := fmt.Sprintf("%s@telegram.user", username)
	if username == "" {
		email = fmt.Sprintf("tg%s@telegram.user", userID)
	}

	user := goth.User{
		Provider:  p.name,
		UserID:    userID,
		Name:      name,
		NickName:  username,
		FirstName: firstName,
		LastName:  lastName,
		AvatarURL: photoURL,
		Email:     email,
	}

	rawData, _ := json.Marshal(data)
	user.RawData = map[string]interface{}{"telegram": string(rawData)}

	return user, nil
}

// verifyAuthData verifies the Telegram auth data hash against the bot token.
func (p *TelegramProvider) verifyAuthData(data map[string]interface{}) error {
	hash, ok := data["hash"].(string)
	if !ok {
		return fmt.Errorf("missing hash in auth data")
	}

	var keys []string
	for k := range data {
		if k != "hash" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var checkStrings []string
	for _, k := range keys {
		checkStrings = append(checkStrings, fmt.Sprintf("%s=%v", k, data[k]))
	}
	checkString := strings.Join(checkStrings, "\n")

	secretKey := sha256.Sum256([]byte(p.botToken))

	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(checkString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(hash), []byte(expectedHash)) {
		return fmt.Errorf("hash mismatch")
	}

	return nil
}

// RefreshToken is not available for Telegram.
func (p *TelegramProvider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	return nil, fmt.Errorf("refresh token not available for telegram")
}

// RefreshTokenAvailable returns false for Telegram.
func (p *TelegramProvider) RefreshTokenAvailable() bool {
	return false
}

// UnmarshalSession deserializes a Telegram session from string.
func (p *TelegramProvider) UnmarshalSession(data string) (goth.Session, error) {
	sess := &TelegramSession{}
	err := json.Unmarshal([]byte(data), sess)
	return sess, err
}

// TelegramSession implements goth.Session for Telegram.
type TelegramSession struct {
	AuthURL  string                 `json:"auth_url"`
	AuthData map[string]interface{} `json:"auth_data"`
}

// GetAuthURL returns the authorization URL.
func (s *TelegramSession) GetAuthURL() (string, error) {
	return s.AuthURL, nil
}

// Authorize processes the callback parameters for Telegram.
// For Telegram, the auth data comes as query parameters to the callback URL.
func (s *TelegramSession) Authorize(provider goth.Provider, params goth.Params) (string, error) {
	if s.AuthData == nil {
		s.AuthData = make(map[string]interface{})
	}

	fields := []string{"id", "first_name", "last_name", "username", "photo_url", "auth_date", "hash"}
	for _, key := range fields {
		val := params.Get(key)
		if val == "" {
			continue
		}
		if key == "id" || key == "auth_date" {
			if intVal, err := strconv.ParseInt(val, 10, 64); err == nil {
				s.AuthData[key] = intVal
				continue
			}
		}
		s.AuthData[key] = val
	}

	return "", nil
}

// Marshal serializes the session to a string.
func (s *TelegramSession) Marshal() string {
	data, _ := json.Marshal(s)
	return string(data)
}
