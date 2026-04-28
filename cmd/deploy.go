package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "One-click production deployment via Docker Compose",
	Long: `Build Docker images and start all PureCore services (PostgreSQL, backend, frontend).

This command runs the scripts/deploy.sh script which handles:
  - Checking prerequisites (Docker, Docker Compose v2)
  - Auto-copying .env.example → .env if needed
  - Validating required secrets (DB_PASSWORD, JWT_SECRET)
  - Building optimized backend and frontend Docker images
  - Starting all services with health checks

Usage:
  purecore deploy              # Full build & deploy
  purecore deploy --build-only # Build images only
  purecore deploy --start-only # Start existing containers
  purecore deploy --down       # Stop all services
  purecore deploy --status     # Show service status`,
	Run: deployRun,
}

func init() {
	rootCmd.AddCommand(deployCmd)
}

func deployRun(cmd *cobra.Command, args []string) {
	// Build arguments to pass to the deploy script
	scriptArgs := []string{"scripts/deploy.sh"}
	if len(args) > 0 {
		scriptArgs = append(scriptArgs, args...)
	}

	// Execute the deploy script
	command := exec.Command("bash", scriptArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	if err := command.Run(); err != nil {
		fmt.Printf("✗ Deploy failed: %v\n", err)
		os.Exit(1)
	}
}
