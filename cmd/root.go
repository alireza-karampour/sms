/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/spf13/cobra"
)

var (
	Nats *string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "sms",
	Short: "a minimal SMS gateway",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := nats.Connect(fmt.Sprintf("nats://%s", *Nats))
		if err != nil {
			return err
		}
		js, err := jetstream.New(conn)
		if err != nil {
			return err
		}
		jss, err := js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
			Name:        "sms",
			Description: "testing",
			Subjects:    []string{"SmsRequest"},
			Storage:     jetstream.FileStorage,
		})
		if err != nil {
			return err
		}
		msg := nats.NewMsg("SmsRequest")
		msg.Data = []byte("Hello")

		ack, err := js.PublishMsg(context.Background(), msg)
		if err != nil {
			return err
		}
		fmt.Printf("seq: %d\n", ack.Sequence)

		rawMsg, err := jss.GetMsg(context.Background(), ack.Sequence, jetstream.WithGetMsgSubject("SmsRequest"))
		if err != nil {
			return err
		}
		fmt.Printf("GotMsg: %s\n", string(rawMsg.Data))

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	Nats = rootCmd.Flags().StringP("nats", "n", "", "nats url")
	err := rootCmd.MarkFlagRequired("nats")
	if err != nil {
		panic(err)
	}
}
