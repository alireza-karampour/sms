package nats

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

func Connect(addr string) (*nats.Conn, error) {
	nc, err := nats.Connect(fmt.Sprintf("nats://%s", addr))
	if err != nil {
		return nil, err
	}
	return nc, nil
}
