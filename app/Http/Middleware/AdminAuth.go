package middleware

import (
	"os"
	"strings"

	"purecore/core"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// JWT secret for signing admin tokens
func jwtSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "purecore-admin-secret-change-in-production"
	}
	return secret
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

		// Store admin info in Locals for downstream handlers
		c.Locals("admin_user_id", uint(claims["user_id"].(float64)))
		c.Locals("admin_username", claims["username"].(string))
		// Also store user for compatibility with existing code
		c.Locals("user", map[string]string{
			"id":   "",
			"name": claims["username"].(string),
		})

		return c.Next()
	}
}

// GenerateAdminToken creates a JWT token for an admin user
func GenerateAdminToken(userID uint, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
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
