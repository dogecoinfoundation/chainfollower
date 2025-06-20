package chainfollower

import (
	"testing"

	"github.com/dogecoinfoundation/chainfollower/pkg/messages"
	"github.com/dogecoinfoundation/chainfollower/pkg/rpc"
	"github.com/dogecoinfoundation/chainfollower/pkg/state"
	"github.com/dogecoinfoundation/chainfollower/pkg/types"
)

func TestBlockMessageReceived(t *testing.T) {
	testTransport := rpc.NewTestRpcTransport()

	t.Log("Waiting for message XXX")

	follower := NewChainFollower(testTransport)

	initialChainPos := &state.ChainPos{
		BlockHash:   "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691",
		BlockHeight: 0,
	}

	testTransport.AddBlockAndHeader(&types.Block{
		Hash:          "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691",
		Confirmations: 1,
	}, &types.BlockHeader{
		Hash:          "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691",
		NextBlockHash: "0000000000000000000000000000000000000000000000000000000000000000",
	})

	testTransport.AddBlockAndHeader(&types.Block{
		Hash:          "0000000000000000000000000000000000000000000000000000000000000000",
		Confirmations: 1,
	}, &types.BlockHeader{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	})

	chainPos, err := follower.FetchStartingPos(initialChainPos)
	if err != nil {
		t.Errorf("Error fetching starting pos: %v", err)
	}

	rawMessage, err := follower.GetNextMessage(chainPos)
	if err != nil {
		t.Errorf("Error fetching next message: %v", err)
	}

	if rawMessage == nil {
		t.Errorf("Block message received is nil")
	}

	blockMessage := rawMessage.(messages.BlockMessage)
	if blockMessage.Block.Hash != "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691" {
		t.Errorf("Block message received is not correct")
	}

	rawMessage, err = follower.GetNextMessage(rawMessage.(messages.BlockMessage).ChainPos)
	if err != nil {
		t.Errorf("Error fetching next message: %v", err)
	}
	if rawMessage == nil {
		t.Errorf("Block message received is nil")
	}

	blockMessage = rawMessage.(messages.BlockMessage)
	if blockMessage.Block.Hash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Block message received is not correct")
	}
}

func TestRollbackMessage(t *testing.T) {
	testTransport := rpc.NewTestRpcTransport()

	t.Log("Waiting for message XXX")

	follower := NewChainFollower(testTransport)

	initialChainPos := &state.ChainPos{
		BlockHash:   "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691",
		BlockHeight: 0,
	}

	testTransport.AddBlockAndHeader(&types.Block{
		Hash: "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691",
	}, &types.BlockHeader{
		Hash:          "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691",
		NextBlockHash: "0000000000000000000000000000000000000000000000000000000000000000",
		Confirmations: 1,
	})

	testTransport.AddBlockAndHeader(&types.Block{
		Hash: "0000000000000000000000000000000000000000000000000000000000000000",
	}, &types.BlockHeader{
		Hash:          "0000000000000000000000000000000000000000000000000000000000000000",
		NextBlockHash: "1111111111111111111111111111111111111111111111111111111111111111",
		Confirmations: 1,
	})

	testTransport.AddBlockAndHeader(&types.Block{
		Hash: "1111111111111111111111111111111111111111111111111111111111111111",
	}, &types.BlockHeader{
		Hash:              "1111111111111111111111111111111111111111111111111111111111111111",
		PreviousBlockHash: "0000000000000000000000000000000000000000000000000000000000000000",
		Confirmations:     -1,
	})

	chainPos, err := follower.fetchStartingPos(initialChainPos)
	if err != nil {
		t.Errorf("Error fetching starting pos: %v", err)
	}

	first, err := follower.GetNextMessage(chainPos)
	if err != nil {
		t.Errorf("Error fetching next message: %v", err)
	}

	second, err := follower.GetNextMessage(first.(messages.BlockMessage).ChainPos)
	if err != nil {
		t.Errorf("Error fetching next message: %v", err)
	}

	rollback, err := follower.GetNextMessage(second.(messages.BlockMessage).ChainPos)
	if err != nil {
		t.Errorf("Error fetching next message: %v", err)
	}

	secondAgain, err := follower.GetNextMessage(rollback.(messages.RollbackMessage).NewChainPos)
	if err != nil {
		t.Errorf("Error fetching next message: %v", err)
	}

	if first.(messages.BlockMessage).Block.Hash != "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691" {
		t.Errorf("Block message received is not correct")
	}

	if second.(messages.BlockMessage).Block.Hash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Block message received is not correct")
	}

	if rollback.(messages.RollbackMessage).NewChainPos.BlockHash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Rollback message received is not correct")
	}

	if rollback.(messages.RollbackMessage).OldChainPos.BlockHash != "1111111111111111111111111111111111111111111111111111111111111111" {
		t.Errorf("Rollback message received is not correct")
	}

	if secondAgain.(messages.BlockMessage).Block.Hash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Block message received is not correct")
	}
}
