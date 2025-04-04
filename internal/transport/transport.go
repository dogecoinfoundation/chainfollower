package transport

import "dogecoin.org/chainfollower/internal/types"

type Transport interface {
	GetBlock(hash string) (*types.Block, error)
	GetBlockHeader(hash string) (*types.BlockHeader, error)
	GetBlockCount() (int64, error)
	GetBestBlockHash() (string, error)
	GetBlockchainInfo() (*types.BlockchainInfo, error)
	GetBlockHash(height int64) (string, error)
}
