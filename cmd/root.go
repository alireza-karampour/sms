package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	viper.SetConfigName("SmsGW")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config")
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Errorf("viper failed to read config: %s", err)
		os.Exit(1)
	}
	logrus.Info("config file read")
}
