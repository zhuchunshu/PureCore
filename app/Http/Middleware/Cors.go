package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func Cors() fiber.Handler {
	// Read allowed origins from CORS_ALLOW_ORIGINS env var (comma-separated)
	// Defaults to "*" for development convenience
	originsStr := os.Getenv("CORS_ALLOW_ORIGINS")
	if originsStr == "" {
		originsStr = "*"
	}
	var origins []string
	if originsStr == "*" {
		origins = []string{"*"}
	} else {
		origins = strings.Split(originsStr, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
	}
	return cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH"},
		AllowCredentials: originsStr != "*", // credentials only when specific origins
	})
}
