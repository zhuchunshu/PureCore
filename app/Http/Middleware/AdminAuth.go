package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	models "purecore/app/Models"
	"purecore/core"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// jwtSecret returns the JWT secret from the centralized Config system.
// Falls back to a default value for development if not configured.
func jwtSecret() string {
	secret := core.GetConfig().String("JWT_SECRET")
	if secret == "" {
		// In development, allow a default for convenience; production must set JWT_SECRET.
		if core.GetConfig().IsProduction() {
			panic("JWT_SECRET environment variable is required in production")
		}
		secret = "purecore-dev-secret-do-not-use-in-production"
	}
	return secret
}

// accessTokenExpiry returns the access token expiration duration
func accessTokenExpiry() time.Duration {
	return 15 * time.Minute
}

// refreshTokenExpiry returns the refresh token expiration duration
func refreshTokenExpiry() time.Duration {
	return 7 * 24 * time.Hour
}

// GenerateRefreshToken creates a cryptographically random refresh token
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// AdminAuth returns a JWT middleware for admin authentication.
// Validates the Bearer token from the Authorization header, extracts
// user_id and username claims, and stores them in Locals.
func AdminAuth() fiber.Handler {
	return func(c fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			res := core.NewResponse(c)
			return res.Unauthorized()
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims,
			func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret()), nil
			},
		)
		if err != nil || !token.Valid {
			res := core.NewResponse(c)
			return res.Unauthorized()
		}

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			res := core.NewResponse(c)
			return res.Unauthorized()
		}
		userID := uint(userIDFloat)
		username, ok := claims["username"].(string)
		if !ok {
			res := core.NewResponse(c)
			return res.Unauthorized()
		}

		// Verify token_version matches the database
		if tv, ok := claims["token_version"]; ok {
			var admin models.AdminUser
			if err := core.DB().First(&admin, userID).Error; err != nil {
				res := core.NewResponse(c)
				return res.Unauthorized()
			}
			tvFloat, ok := tv.(float64)
			if !ok || int(tvFloat) != admin.TokenVersion {
				res := core.NewResponse(c)
				return res.Unauthorized()
			}
		}

		// Store admin info in Locals for downstream handlers
		c.Locals("admin_user_id", userID)
		c.Locals("admin_username", username)
		// Also store user for compatibility with existing code
		c.Locals("user", map[string]string{
			"id":   "",
			"name": username,
		})

		return c.Next()
	}
}

// GenerateAdminToken creates a JWT access token for an admin user (short-lived)
func GenerateAdminToken(userID uint, username string, tokenVersion int) (string, error) {
	claims := jwt.MapClaims{
		"user_id":       userID,
		"username":      username,
		"token_version": tokenVersion,
		"exp":           time.Now().Add(accessTokenExpiry()).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret()))
}

// GetAdminUserID extracts the admin user ID from the JWT token in the request
func GetAdminUserID(c fiber.Ctx) uint {
	if id := c.Locals("admin_user_id"); id != nil {
		return id.(uint)
	}
	return 0
}

// GetAdminUsername extracts the admin username from the JWT token in the request
func GetAdminUsername(c fiber.Ctx) string {
	if name := c.Locals("admin_username"); name != nil {
		return name.(string)
	}
	return ""
}
