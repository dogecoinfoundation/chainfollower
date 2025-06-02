package chainfollower

import (
	"testing"

	"dogecoin.org/chainfollower/pkg/messages"
	"dogecoin.org/chainfollower/pkg/rpc"
	"dogecoin.org/chainfollower/pkg/state"
	"dogecoin.org/chainfollower/pkg/types"
)

func TestBlockMessageReceived(t *testing.T) {
	testTransport := rpc.NewTestRpcTransport()

	t.Log("Waiting for message XXX")

	follower := NewChainFollower(testTransport)

	chainPos := &state.ChainPos{
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

	messageChan := follower.Start(chainPos)

	rawMessage := <-messageChan
	if rawMessage == nil {
		t.Errorf("Block message received is nil")
	}

	blockMessage := rawMessage.(messages.BlockMessage)
	if blockMessage.Block.Hash != "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691" {
		t.Errorf("Block message received is not correct")
	}

	rawMessage = <-messageChan
	if rawMessage == nil {
		t.Errorf("Block message received is nil")
	}

	blockMessage = rawMessage.(messages.BlockMessage)
	if blockMessage.Block.Hash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Block message received is not correct")
	}

	if len(messageChan) != 0 {
		t.Errorf("Shouldn't be messages if we are waiting on Next Hash")
	}
}

func TestRollbackMessage(t *testing.T) {
	testTransport := rpc.NewTestRpcTransport()

	t.Log("Waiting for message XXX")

	follower := NewChainFollower(testTransport)

	chainPos := &state.ChainPos{
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

	messageChan := follower.Start(chainPos)

	first := (<-messageChan).(messages.BlockMessage)
	second := (<-messageChan).(messages.BlockMessage)
	rollback := (<-messageChan).(messages.RollbackMessage)
	secondAgain := (<-messageChan).(messages.BlockMessage)

	if first.Block.Hash != "1a91e3dace36e2be3bf030a65679fe821aa1d6ef92e7c9902eb318182c355691" {
		t.Errorf("Block message received is not correct")
	}

	if second.Block.Hash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Block message received is not correct")
	}

	if rollback.NewChainPos.BlockHash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Rollback message received is not correct")
	}

	if rollback.OldChainPos.BlockHash != "1111111111111111111111111111111111111111111111111111111111111111" {
		t.Errorf("Rollback message received is not correct")
	}

	if secondAgain.Block.Hash != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Block message received is not correct")
	}
}
