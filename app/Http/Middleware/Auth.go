package middleware

import (
	"fmt"
	"strings"

	"purecore/core"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Auth returns a JWT middleware for regular user authentication.
// It validates the Bearer token from the Authorization header and extracts
// user claims into Locals.
func Auth() fiber.Handler {
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

		// Store user info in Locals for downstream handlers
		c.Locals("user", map[string]string{
			"id":   fmt.Sprintf("%.0f", claims["user_id"]),
			"name": getClaimsString(claims, "username"),
		})

		return c.Next()
	}
}

// GenerateUserToken creates a JWT token for a regular user
func GenerateUserToken(userID uint, username string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret()))
}

func getClaimsString(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key]; ok {
		if s, ok := val.(string); ok {
			return s
		}
	}
	return ""
}
