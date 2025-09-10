package api

import (
	"context"
	"fmt"

	. "github.com/alireza-karampour/sms/cmd"
	"github.com/alireza-karampour/sms/internal/controllers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	UserController        *controllers.User
	PhoneNumberController *controllers.PhoneNumber
)

// ApiCmd represents the api command
var ApiCmd = &cobra.Command{
	Use:   "api",
	Short: "runs the REST Api server",
	RunE: func(cmd *cobra.Command, args []string) error {
		dbConn, err := pgx.Connect(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%d",
			viper.GetString("api.postgres.username"),
			viper.GetString("api.postgres.password"),
			viper.GetString("api.postgres.address"),
			viper.GetInt("api.postgres.port"),
		))
		if err != nil {
			return err
		}

		r := gin.Default()
		root := r.Group("/")
		UserController = controllers.NewUser(root, dbConn)
		PhoneNumberController = controllers.NewPhoneNumber(root, dbConn)

		return r.Run(viper.GetString("api.listen"))
	},
}

func init() {
	RootCmd.AddCommand(ApiCmd)
	ApiCmd.Flags().StringP("nats", "n", "127.0.0.1:4222", "nats url")
	ApiCmd.Flags().StringP("listen", "l", "0.0.0.0:8080", "address to listen on")

	viper.BindPFlag("nats", ApiCmd.Flags().Lookup("nats"))
	viper.BindPFlag("listen", ApiCmd.Flags().Lookup("listen"))
}
