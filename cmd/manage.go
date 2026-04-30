package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"purecore/manage"

	"github.com/spf13/cobra"
)

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Interactive management TUI for PureCore",
	Long: `Launch the PureCore management TUI to:
  - Switch admin route prefix
  - Switch frontend theme
  - Update PureCore version
  - View service status
  - Restart services / View logs
  - Open shell in containers

Requires docker compose and a valid PureCore installation
in the current directory.`,
	RunE: manageRun,
}

var (
	manageLogs  string
	manageShell string
)

func init() {
	manageCmd.Flags().StringVar(&manageLogs, "logs", "", "View logs for a service (all, backend, frontend, database)")
	manageCmd.Flags().StringVar(&manageShell, "shell", "", "Open shell in a container (backend, frontend, db)")
	rootCmd.AddCommand(manageCmd)
}

func manageRun(cmd *cobra.Command, args []string) error {
	// Handle --logs flag (non-interactive)
	if manageLogs != "" {
		valid, composeFile, _ := manage.CheckInstallation()
		if !valid {
			return fmt.Errorf("not a PureCore project directory")
		}
		return manage.RunLogs(composeFile, manageLogs)
	}

	// Handle --shell flag (non-interactive)
	if manageShell != "" {
		containerMap := map[string]string{
			"backend":  "purecore-backend",
			"frontend": "purecore-frontend",
			"db":       "purecore-db",
		}
		container, ok := containerMap[manageShell]
		if !ok {
			return fmt.Errorf("unknown container: %s (valid: backend, frontend, db)", manageShell)
		}
		// Try /bin/sh first, then /bin/bash
		c := exec.Command("docker", "exec", "-it", container, "/bin/sh")
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			c = exec.Command("docker", "exec", "-it", container, "/bin/bash")
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		}
		return nil
	}

	// Launch interactive TUI
	return manage.Run()
}
