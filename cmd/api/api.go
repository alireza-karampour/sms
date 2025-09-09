package cmd

import (
	"fmt"
	. "github.com/alireza-karampour/sms/cmd"
	"github.com/spf13/cobra"
)

var (
	Nats *string
)

// apiCmd represents the api command
var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "runs the REST Api server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("api called")
	},
}

func init() {
	RootCmd.AddCommand(apiCmd)

	Nats = RootCmd.Flags().StringP("nats", "n", "", "nats url")
	err := RootCmd.MarkFlagRequired("nats")
	if err != nil {
		panic(err)
	}
}
