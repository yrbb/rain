package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		testHandler{}.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	// testCmd.PersistentFlags().String("foo", "", "A help for foo")
	// testCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type testHandler struct{}

func (g testHandler) Run(cmd *cobra.Command, args []string) {
	fmt.Println("test called")
}
