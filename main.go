package main

import (
	"github.com/alireza-karampour/sms/cmd"
	_ "github.com/alireza-karampour/sms/cmd/api"
)

func main() {
	cmd.Execute()
}
