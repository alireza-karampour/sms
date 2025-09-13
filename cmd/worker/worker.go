package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	. "github.com/alireza-karampour/sms/cmd"
	"github.com/alireza-karampour/sms/internal/workers"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Worker *workers.Sms
)

// WorkerCmd represents the worker command
var WorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "starts worker node for sms request handling",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
		defer cancel()
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{
			ForceColors:            true,
			DisableLevelTruncation: true,
		})
		pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgresql://%s:%s@%s:%d",
			viper.GetString("worker.postgres.username"),
			viper.GetString("worker.postgres.password"),
			viper.GetString("worker.postgres.address"),
			viper.GetInt("worker.postgres.port"),
		))
		if err != nil {
			return err
		}
		err = pool.Ping(context.Background())
		if err != nil {
			return err
		}

		natsAddress := viper.GetString("worker.nats.address")
		Worker, err = workers.NewSms(ctx, natsAddress, pool)
		if err != nil {
			return err
		}
		err = Worker.Start(ctx)
		if err != nil {
			return err
		}

		<-ctx.Done()
		return nil
	},
}

func init() {
	RootCmd.AddCommand(WorkerCmd)
}
