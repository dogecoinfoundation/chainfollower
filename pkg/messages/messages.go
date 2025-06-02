package messages

import (
	"github.com/dogecoinfoundation/chainfollower/pkg/state"
	"github.com/dogecoinfoundation/chainfollower/pkg/types"
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
