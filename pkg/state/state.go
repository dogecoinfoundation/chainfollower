package state

type ChainPos struct {
	BlockHash          string // last block processed ("" at genesis)
	BlockHeight        int64  // height of last block (0 at genesis)
	WaitingForNextHash bool   // if the block has been fetched
}
