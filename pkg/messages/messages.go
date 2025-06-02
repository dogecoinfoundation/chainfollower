package messages

import (
	"dogecoin.org/chainfollower/pkg/state"
	"dogecoin.org/chainfollower/pkg/types"
)

type Message interface {
}

type BlockMessage struct {
	Message
	Block    *types.Block
	ChainPos *state.ChainPos
}

type RollbackMessage struct {
	Message
	OldChainPos *state.ChainPos
	NewChainPos *state.ChainPos
}
