package cmd

import (
	"log"

	middleware "purecore/app/Http/Middleware"
	providers "purecore/app/Providers"
	"purecore/core"
	_ "purecore/database/migrations"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the PureCore HTTP server",
	Long:  `Start the PureCore HTTP server on the configured port (default: 9002).`,
	Run:   serveRun,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serveRun(cmd *cobra.Command, args []string) {
	// Create and bootstrap the application using the new Application container.
	// This centralizes config, database, language, routing, and service providers.
	app := core.NewApplication().
		AddProviders(&providers.RouteServiceProvider{})

	if err := app.RunWithMiddleware(
		middleware.Cors(),
		middleware.Lang(),
	); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
