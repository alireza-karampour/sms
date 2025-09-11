package worker

import (
	"context"
	"os"
	"os/signal"

	. "github.com/alireza-karampour/sms/cmd"
	"github.com/alireza-karampour/sms/internal/workers"
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

		natsAddress := viper.GetString("worker.nats.address")
		Worker, err = workers.NewSms(ctx, natsAddress)
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
