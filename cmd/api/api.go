package cmd

import (
	. "github.com/alireza-karampour/sms/cmd"
	"github.com/alireza-karampour/sms/internal/controllers"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var (
	Nats   *string
	Listen *string
)

var (
	UserController *controllers.User
)

// ApiCmd represents the api command
var ApiCmd = &cobra.Command{
	Use:   "api",
	Short: "runs the REST Api server",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := gin.Default()
		root := r.Group("/")

		UserController = controllers.NewUser(root)

		return r.Run(*Listen)
	},
}

func init() {
	RootCmd.AddCommand(ApiCmd)
	Nats = ApiCmd.Flags().StringP("nats", "n", "127.0.0.1:4222", "nats url")
	Listen = ApiCmd.Flags().StringP("listen", "l", "0.0.0.0:8080", "address to listen on")
}
