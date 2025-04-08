package main

import (
	"log"

	"dogecoin.org/chainfollower/pkg/chainfollower"
	"dogecoin.org/chainfollower/pkg/config"
	"dogecoin.org/chainfollower/pkg/messages"
	"dogecoin.org/chainfollower/pkg/rpc"
	"dogecoin.org/chainfollower/pkg/store"
)

func main() {
	config, err := config.LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	rpcClient := rpc.NewRpcTransport(config)
	chainfollower := chainfollower.NewChainFollower(rpcClient)

	chainPos, err := store.LoadChainPos("position.json")
	if err != nil {
		log.Fatal(err)
	}

	messageChan := chainfollower.Start(chainPos)

	for message := range messageChan {
		switch msg := message.(type) {
		case messages.BlockMessage:
			log.Println("Received message from chainfollower:")
			log.Println(msg.Block)
			log.Println(msg.ChainPos)

			store.SaveChainPos("data.json", msg.ChainPos)
		case messages.RollbackMessage:
			log.Println("Received rollback message from chainfollower:")
			log.Println(msg.OldChainPos)
			log.Println(msg.NewChainPos)

			store.SaveChainPos("data.json", msg.NewChainPos)
		default:
			log.Println("Received unknown message from chainfollower:")
		}
	}
}
