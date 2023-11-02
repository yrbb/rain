package cmd

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "rain-cli",
		Short: "A generator for Rain based Applications",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(initCmd)
	// rootCmd.AddCommand(routerCmd)
}
