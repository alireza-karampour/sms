package main

import (
	"github.com/alireza-karampour/sms/cmd"
	_ "github.com/alireza-karampour/sms/cmd/api"
	_ "github.com/alireza-karampour/sms/cmd/worker"
)

func main() {
	cmd.Execute()
}
