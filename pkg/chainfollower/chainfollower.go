package chainfollower

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dogecoin.org/chainfollower/internal/commands"
	"dogecoin.org/chainfollower/internal/doge"
	"dogecoin.org/chainfollower/internal/messages"
	"dogecoin.org/chainfollower/internal/state"
	"dogecoin.org/chainfollower/internal/transport"
)

const (
	RETRY_DELAY        = 5 * time.Second        // for RPC and Database errors.
	WRONG_CHAIN_DELAY  = 5 * time.Minute        // for "Wrong Chain" error (essentially stop)
	WAIT_INITIAL_BLOCK = 30 * time.Second       // for Initial Block Download
	CONFLICT_DELAY     = 250 * time.Millisecond // for Database conflicts (concurrent transactions)
	BLOCKS_PER_COMMIT  = 10                     // number of blocks per database commit.
)

type ChainFollower struct {
	transport          transport.Transport
	chain              *doge.ChainParams
	Commands           chan any                         // receive ReSyncChainFollowerCmd etc.
	confirmations      int                              // required number of block confirmations.
	stopping           bool                             // set to exit the main loop.
	SetSync            *commands.ReSyncChainFollowerCmd // pending ReSync command.
	Messages           chan messages.Message            // send messages to the main loop.
	MessageChannelSize int

	// receive signals from the main loop.
}

func NewChainFollower(transport transport.Transport) *ChainFollower {
	return &ChainFollower{transport: transport, MessageChannelSize: 0}
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
		for {
			select {
			case sig := <-sigCh: // sigterm/sigint caught
				fmt.Printf("Caught %v signal, shutting down\n", sig)
				c.Stop()
				continue
			}
		}
	}()
}

func (c *ChainFollower) serviceMain(chainState *state.ChainPos) {
	chainPos, err := c.FetchStartingPos(chainState)

	if err != nil {
		log.Println("ChainFollower: FetchStartingPos failed:", err)
		return
	}

	for {
		blockHeader, err := c.transport.GetBlockHeader(chainPos.BlockHash)
		if err != nil {
			log.Println("ChainFollower: GetBlockHeader failed:", err)
			return
		}

		if blockHeader.IsOnChain() {
			if !chainPos.WaitingForNextHash {
				fmt.Println("ChainFollower: GetBlock", chainPos.BlockHash)
				block, err := c.transport.GetBlock(blockHeader.Hash)
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
			chainPos, err = c.rollbackToOnChainBlock(blockHeader.PreviousBlockHash, chainPos)
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

// LOOP
// - Fetch Block Header from Position (Position.BlockHash)

// - Block Header still on chain?
//     - Emit 'Block Message'
//     - Update Position to Next Hash
// - Block Header no longer on chain?
//     - Walk back until we find a block is on chain (this is the new height)
//     - Emit 'Rollback Message' with new Block Height/Hash
//     - Update Position
// LOOP

func (c *ChainFollower) rollbackToOnChainBlock(fromHash string, oldPos *state.ChainPos) (*state.ChainPos, error) {
	for {
		// Fetch the block header for the previous block.
		log.Println("ChainFollower: fetching previous header:", fromHash)
		block, err := c.transport.GetBlockHeader(fromHash)
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
			pos := &state.ChainPos{block.Hash, block.Height, false}
			return pos, nil
		}
	}
}

func (c *ChainFollower) FetchStartingPos(initialChainPos *state.ChainPos) (*state.ChainPos, error) {
	// Retry loop for transaction error or wrong-chain error.
	for {
		genesisHash, err := c.transport.GetBlockHash(0)
		if err != nil {
			return nil, err
		}

		chain, err := doge.ChainFromGenesisHash(genesisHash)
		if err != nil {
			log.Println("ChainFollower: UNRECOGNISED CHAIN!")
			log.Println("ChainFollower: Block#0 on Core Node:", genesisHash)
			log.Println("ChainFollower: The Genesis block does not match any of our ChainParams")
			log.Println("ChainFollower: Please connect to a Dogecoin Core Node")
			c.sleepForRetry(nil, WRONG_CHAIN_DELAY)
			continue
		}
		c.chain = chain

		info, err := c.transport.GetBlockchainInfo()
		if err != nil {
			return nil, err
		}

		if info.InitialBlockDownload {
			log.Println("ChainFollower: waiting for Core initial block download")
			c.sleepForRetry(nil, WAIT_INITIAL_BLOCK)
			continue
		}

		if initialChainPos.BlockHash != "" {
			log.Println("ChainFollower: RESUME SYNC :", initialChainPos.BlockHeight)

			return &state.ChainPos{initialChainPos.BlockHash, initialChainPos.BlockHeight, false}, nil
		} else {
			firstHeight, err := c.transport.GetBlockCount()
			if err != nil {
				return nil, err
			}

			if firstHeight > 100 {
				firstHeight -= 100
			} else {
				firstHeight = 0
			}

			firstBlockHash, err := c.transport.GetBlockHash(firstHeight)
			if err != nil {
				return nil, err
			}

			return &state.ChainPos{firstBlockHash, firstHeight, false}, nil
		}
	}
}

func (c *ChainFollower) sleepForRetry(err error, delay time.Duration) {
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
