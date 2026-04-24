package middleware

import (
	"strings"

	"purecore/core"

	"github.com/gofiber/fiber/v3"
)

func Lang() fiber.Handler {
	return func(c fiber.Ctx) error {
		acceptLang := c.Get("Accept-Language")
		locale := "zh" // default Chinese
		if acceptLang != "" {
			parts := strings.Split(acceptLang, ",")
			if len(parts) > 0 {
				primaryLang := strings.Split(strings.TrimSpace(parts[0]), "-")[0]
				if primaryLang != "" {
					locale = primaryLang
				}
			}
		}
		c.Locals("locale", locale)
		core.GetLang().SetLocale(locale)
		return c.Next()
	}
}
