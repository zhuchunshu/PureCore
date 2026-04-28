package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var makeMiddlewareCmd = &cobra.Command{
	Use:   "make:middleware [name]",
	Short: "Create a new middleware file",
	Long: `Create a new middleware file in app/Http/Middleware/.
This generates a middleware that can be used in route definitions.
The middleware can be registered as a named middleware for use with
router.MiddlewareByName("name") or used directly.`,
	Args: cobra.ExactArgs(1),
	Run:  makeMiddlewareRun,
}

func init() {
	rootCmd.AddCommand(makeMiddlewareCmd)
}

func makeMiddlewareRun(cmd *cobra.Command, args []string) {
	name := args[0]
	// Ensure first letter is uppercase for consistency
	name = strings.ToUpper(name[:1]) + name[1:]

	fileName := name + ".go"
	filePath := filepath.Join("app", "Http", "Middleware", fileName)

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(filePath), 0755)

	lowerName := strings.ToLower(name)
	content := fmt.Sprintf(`package middleware

import "github.com/gofiber/fiber/v3"

// %s returns a middleware that %s.
// Register this middleware in a RouteServiceProvider or use it directly:
//   router.Middleware(%s())
func %s() fiber.Handler {
	return func(c fiber.Ctx) error {
		// Add your middleware logic here

		// Call the next handler
		return c.Next()
	}
}
`, name, lowerName, name, name)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("✗ Failed to create middleware: %v\n", err)
		return
	}

	fmt.Printf("✓ Middleware created: %s\n", filePath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Add middleware.%s() to your route definitions\n", name)
	fmt.Println("  2. Or register it as a named middleware in a service provider")
	fmt.Printf("  3. For global middleware, add it to RunWithMiddleware() in cmd/serve.go\n")
}
