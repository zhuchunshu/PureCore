package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "purecore",
	Short: "PureCore - A full-stack Go web development framework",
	Long: `PureCore is a full-stack Go web development framework that wraps GoFiber v3 
into a Laravel-like development style. Use the 'serve' subcommand to start the server.`,
}

func Execute() error {
	return rootCmd.Execute()
}
