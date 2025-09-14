package api

import (
	"context"
	"fmt"

	. "github.com/alireza-karampour/sms/cmd"
	"github.com/alireza-karampour/sms/internal/controllers"
	"github.com/alireza-karampour/sms/pkg/nats"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
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
		pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable",
			viper.GetString("api.postgres.username"),
			viper.GetString("api.postgres.password"),
			viper.GetString("api.postgres.address"),
			viper.GetInt("api.postgres.port"),
			viper.GetString("api.postgres.username"),
		))
		if err != nil {
			return err
		}
		err = pool.Ping(context.Background())
		if err != nil {
			return err
		}

		natsConn, err := nats.Connect(viper.GetString("api.nats.address"))
		if err != nil {
			return err
		}

		r := gin.Default()

		// Add health check endpoint
		r.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "healthy",
				"service": "sms-api",
			})
		})

		root := r.Group("/")
		UserController = controllers.NewUser(root, pool)
		PhoneNumberController = controllers.NewPhoneNumber(root, pool)
		SmsController, err = controllers.NewSms(root, pool, natsConn)
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
