package worker

import (
	. "github.com/alireza-karampour/sms/cmd"
	"github.com/alireza-karampour/sms/pkg/nats"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// WorkerCmd represents the worker command
var WorkerCmd = &cobra.Command{
	Use:   "worker",
	Short: "starts worker node for sms request handling",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := nats.Connect(viper.GetString("worker.nats.address"))
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(WorkerCmd)
}
