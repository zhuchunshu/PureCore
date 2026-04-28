package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var makeModuleCmd = &cobra.Command{
	Use:   "make:module [name]",
	Short: "Create a new service provider module",
	Long: `Create a new service provider module in app/Providers/. 
This generates a new provider that implements the core.ServiceProvider interface.
The module can register its own routes, middleware, and other services.`,
	Args: cobra.ExactArgs(1),
	Run:  makeModuleRun,
}

func init() {
	rootCmd.AddCommand(makeModuleCmd)
}

func makeModuleRun(cmd *cobra.Command, args []string) {
	name := args[0]
	// Ensure first letter is uppercase for consistency
	name = strings.ToUpper(name[:1]) + name[1:]

	providerName := name + "ServiceProvider"
	fileName := providerName + ".go"
	filePath := filepath.Join("app", "Providers", fileName)

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(filePath), 0755)

	lowerName := strings.ToLower(name)
	content := fmt.Sprintf(`package providers

import (
	"purecore/core"
)

// %s registers %s-related routes and services.
// This provider is auto-registered in cmd/serve.go via app.AddProviders().
type %s struct{}

// Name returns the unique identifier for this provider
func (p *%s) Name() string {
	return "%s"
}

// Register sets up all routes and middleware for this module.
// Add your routes here using the provided router.
func (p *%s) Register(router *core.Router) error {
	// Example: r.Prefix("/api/v1").Group(func(r *core.Router) {
	//     r.Get("/%s", core.H(p.Index))
	// })
	return nil
}

// Boot is called after all providers have been registered.
// Use this for any post-registration initialization.
func (p *%s) Boot() error {
	return nil
}
`, providerName, lowerName, providerName, providerName, lowerName, providerName, lowerName, providerName)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("✗ Failed to create module: %v\n", err)
		return
	}

	fmt.Printf("✓ Module created: %s\n", filePath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Add &providers.%s{} to the AddProviders() chain in cmd/serve.go\n", providerName)
	fmt.Println("  2. Implement your routes in the Register() method")
	fmt.Println("  3. Add any initialization logic in the Boot() method")
}
