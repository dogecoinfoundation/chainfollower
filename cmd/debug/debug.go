package main

import (
	"log"

	"github.com/dogecoinfoundation/chainfollower/pkg/config"
	"github.com/dogecoinfoundation/chainfollower/pkg/rpc"
)

func main() {
	config, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	rpcClient := rpc.NewRpcTransport(config)

	blockCount, err := rpcClient.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(blockCount)

	block, err := rpcClient.GetBlockHash(blockCount)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(block)
}
