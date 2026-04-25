package cmd

import (
	"purecore/core"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run pending database migrations",
	Long:  `Run all registered migrations that have not yet been executed. Uses a migrations table to track execution history.`,
	Run:   migrateRun,
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

func migrateRun(cmd *cobra.Command, args []string) {
	core.InitLang("lang")
	core.RunMigrations()
}
