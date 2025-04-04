package transport

import (
	"dogecoin.org/chainfollower/internal/config"
)

type ZmqTransport struct {
	config *config.Config
}

func NewZmqTransport(config *config.Config) *ZmqTransport {
	return &ZmqTransport{config: config}
}
