package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var ()

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "sms",
	Short: "a minimal SMS gateway",
	RunE: func(cmd *cobra.Command, args []string) error {

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
