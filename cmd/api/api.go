package api

import (
	"context"
	"fmt"

	. "github.com/alireza-karampour/sms/cmd"
	"github.com/alireza-karampour/sms/internal/controllers"
	"github.com/alireza-karampour/sms/pkg/nats"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	UserController        *controllers.User
	PhoneNumberController *controllers.PhoneNumber
	SmsController         *controllers.Sms
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
		natsConn, err := nats.Connect(viper.GetString("api.nats.address"))
		if err != nil {
			return err
		}

		r := gin.Default()
		root := r.Group("/")
		UserController = controllers.NewUser(root, dbConn)
		PhoneNumberController = controllers.NewPhoneNumber(root, dbConn)
		SmsController, err = controllers.NewSms(root, dbConn, natsConn)
		if err != nil {
			return err
		}

		return r.Run(viper.GetString("api.listen"))
	},
}

func init() {
	RootCmd.AddCommand(ApiCmd)

	viper.SetDefault("api.sms.cost", 5)
}
