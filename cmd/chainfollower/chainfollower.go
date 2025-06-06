package main

import (
	"log"

	"github.com/dogecoinfoundation/chainfollower/pkg/chainfollower"
	"github.com/dogecoinfoundation/chainfollower/pkg/config"
	"github.com/dogecoinfoundation/chainfollower/pkg/messages"
	"github.com/dogecoinfoundation/chainfollower/pkg/rpc"
	"github.com/dogecoinfoundation/chainfollower/pkg/store"
)

func main() {
	config, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	rpcClient := rpc.NewRpcTransport(config)
	chainfollower := chainfollower.NewChainFollower(rpcClient)

	initialChainPos, err := store.LoadChainPos("position.json")
	if err != nil {
		log.Fatal(err)
	}

	chainPos, err := chainfollower.FetchStartingPos(initialChainPos)
	if err != nil {
		log.Fatal(err)
	}

	message, err := chainfollower.GetNextMessage(chainPos)
	if err != nil {
		log.Fatal(err)
	}

	for {
		switch msg := message.(type) {
		case messages.BlockMessage:
			log.Println("Received message from chainfollower:")
			log.Println(msg.Block)
			log.Println(msg.ChainPos)

			store.SaveChainPos("data.json", msg.ChainPos)

			chainPos = msg.ChainPos
		case messages.RollbackMessage:
			log.Println("Received rollback message from chainfollower:")
			log.Println(msg.OldChainPos)
			log.Println(msg.NewChainPos)

			store.SaveChainPos("data.json", msg.NewChainPos)

			chainPos = msg.NewChainPos
		default:
			log.Println("Received unknown message from chainfollower:")
		}

		message, err = chainfollower.GetNextMessage(chainPos)
		if err != nil {
			log.Fatal(err)
		}
	}
}
