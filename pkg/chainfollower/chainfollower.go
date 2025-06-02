package chainfollower

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dogecoinfoundation/chainfollower/internal/commands"
	"github.com/dogecoinfoundation/chainfollower/internal/doge"
	"github.com/dogecoinfoundation/chainfollower/pkg/messages"
	"github.com/dogecoinfoundation/chainfollower/pkg/rpc"
	"github.com/dogecoinfoundation/chainfollower/pkg/state"
)

const (
	RETRY_DELAY        = 5 * time.Second        // for RPC and Database errors.
	WRONG_CHAIN_DELAY  = 5 * time.Minute        // for "Wrong Chain" error (essentially stop)
	WAIT_INITIAL_BLOCK = 30 * time.Second       // for Initial Block Download
	CONFLICT_DELAY     = 250 * time.Millisecond // for Database conflicts (concurrent transactions)
	BLOCKS_PER_COMMIT  = 10                     // number of blocks per database commit.
)

type ChainFollower struct {
	rpc                rpc.RpcTransportInterface
	chain              *doge.ChainParams
	Commands           chan any                         // receive ReSyncChainFollowerCmd etc.
	stopping           bool                             // set to exit the main loop.
	SetSync            *commands.ReSyncChainFollowerCmd // pending ReSync command.
	Messages           chan messages.Message            // send messages to the main loop.
	MessageChannelSize int

	// receive signals from the main loop.
}

func NewChainFollower(rpc rpc.RpcTransportInterface) *ChainFollower {
	return &ChainFollower{rpc: rpc, MessageChannelSize: 0}
}

func (c *ChainFollower) Start(chainState *state.ChainPos) chan messages.Message {
	go c.handleSignals()

	c.Messages = make(chan messages.Message, c.MessageChannelSize)

	go c.serviceMain(chainState)

	return c.Messages
}

func (c *ChainFollower) Stop() {
	close(c.Messages)
}

func (c *ChainFollower) handleSignals() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for sig := range sigCh {
			fmt.Printf("Caught %v signal, shutting down\n", sig)
			c.Stop()
		}
	}()
}

func (c *ChainFollower) serviceMain(chainState *state.ChainPos) {
	chainPos, err := c.fetchStartingPos(chainState)

	if err != nil {
		log.Println("ChainFollower: fetchStartingPos failed:", err)
		return
	}

	for {
		blockHeader, err := c.rpc.GetBlockHeader(chainPos.BlockHash)
		if err != nil {
			log.Println("ChainFollower: GetBlockHeader failed:", err)
			return
		}

		if blockHeader.IsOnChain() {
			if !chainPos.WaitingForNextHash {
				// fmt.Println("ChainFollower: GetBlock", chainPos.BlockHash)
				block, err := c.rpc.GetBlock(blockHeader.Hash)
				if err != nil {
					log.Println("ChainFollower: GetBlock failed:", err)
					return
				}

				chainPos.WaitingForNextHash = true

				c.Messages <- messages.BlockMessage{
					Block:    block,
					ChainPos: chainPos,
				}
			}

			chainPos.WaitingForNextHash = blockHeader.NextBlockHash == ""

			if blockHeader.NextBlockHash != "" {
				chainPos.BlockHash = blockHeader.NextBlockHash
				chainPos.BlockHeight = blockHeader.Height
			}

			// TODO : Rethink this
			if chainPos.WaitingForNextHash {
				time.Sleep(1 * time.Second)
			}
		} else {

			oldChainPos := chainPos
			chainPos, err = c.rollbackToOnChainBlock(blockHeader.PreviousBlockHash)
			if err != nil {
				log.Println("ChainFollower: rollbackToOnChainBlock failed:", err)
				return
			}

			oldChainPos.WaitingForNextHash = false
			chainPos.WaitingForNextHash = false

			c.Messages <- messages.RollbackMessage{
				OldChainPos: oldChainPos,
				NewChainPos: chainPos,
			}
		}
	}
}

func (c *ChainFollower) rollbackToOnChainBlock(fromHash string) (*state.ChainPos, error) {
	for {
		// Fetch the block header for the previous block.
		log.Println("ChainFollower: fetching previous header:", fromHash)
		block, err := c.rpc.GetBlockHeader(fromHash)
		if err != nil {
			log.Println("ChainFollower: GetBlockHeader failed:", err)
			return nil, err
		}

		if block.Confirmations == -1 {
			// This block is no longer on-chain, so keep walking backwards.
			fromHash = block.PreviousBlockHash
			// c.checkShutdown() // loops must check for shutdown.
		} else {
			// Found an on-chain block: roll back all chainstate above this block-height.
			pos := &state.ChainPos{
				BlockHash:          block.Hash,
				BlockHeight:        block.Height,
				WaitingForNextHash: false,
			}
			return pos, nil
		}
	}
}

func (c *ChainFollower) fetchStartingPos(initialChainPos *state.ChainPos) (*state.ChainPos, error) {
	// Retry loop for transaction error or wrong-chain error.
	for {
		genesisHash, err := c.rpc.GetBlockHash(0)
		if err != nil {
			return nil, err
		}

		chain, err := doge.ChainFromGenesisHash(genesisHash)
		if err != nil {
			log.Println("ChainFollower: UNRECOGNISED CHAIN!")
			log.Println("ChainFollower: Block#0 on Core Node:", genesisHash)
			log.Println("ChainFollower: The Genesis block does not match any of our ChainParams")
			log.Println("ChainFollower: Please connect to a Dogecoin Core Node")
			c.sleepForRetry(WRONG_CHAIN_DELAY)
			continue
		}
		c.chain = chain

		info, err := c.rpc.GetBlockchainInfo()
		if err != nil {
			return nil, err
		}

		if info.InitialBlockDownload {
			log.Println("ChainFollower: waiting for Core initial block download")
			c.sleepForRetry(WAIT_INITIAL_BLOCK)
			continue
		}

		if initialChainPos.BlockHash != "" {
			log.Println("ChainFollower: RESUME SYNC :", initialChainPos.BlockHeight)

			return &state.ChainPos{
				BlockHash:          initialChainPos.BlockHash,
				BlockHeight:        initialChainPos.BlockHeight,
				WaitingForNextHash: false,
			}, nil
		} else {
			firstHeight, err := c.rpc.GetBlockCount()
			if err != nil {
				return nil, err
			}

			if firstHeight > 100 {
				firstHeight -= 100
			} else {
				firstHeight = 0
			}

			firstBlockHash, err := c.rpc.GetBlockHash(firstHeight)
			if err != nil {
				return nil, err
			}

			return &state.ChainPos{
				BlockHash:          firstBlockHash,
				BlockHeight:        firstHeight,
				WaitingForNextHash: false,
			}, nil
		}
	}
}

func (c *ChainFollower) sleepForRetry(delay time.Duration) {
	if delay == 0 {
		delay = RETRY_DELAY
	}
	select {
	case cmd := <-c.Commands:
		log.Println("ChainFollower: received command")
		switch cm := cmd.(type) {
		case commands.StopChainFollowerCmd:
			c.stopping = true
			panic("stopped") // caught in `Run` method.
		case commands.RestartChainFollowerCmd:
			panic("stopped") // caught in `Run` method.
		case commands.ReSyncChainFollowerCmd:
			c.SetSync = &cm
			panic("restart") // caught in `Run` method.
		default:
			log.Println("ChainFollower: unknown command received (ignored)")
		}
	case <-time.After(delay):
		return
	}
}
