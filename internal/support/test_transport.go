package support

import (
	"fmt"

	"dogecoin.org/chainfollower/internal/transport"
	"dogecoin.org/chainfollower/internal/types"
)

type TestTransport struct {
	transport.Transport
	blocks         []*types.Block
	headers        []*types.BlockHeader
	bestBlockHash  string
	blockCount     int64
	blockChainInfo *types.BlockchainInfo
}

func (t *TestTransport) GetBlock(hash string) (*types.Block, error) {
	for _, block := range t.blocks {
		if block.Hash == hash {
			return block, nil
		}
	}
	return nil, fmt.Errorf("block not found: %s", hash)
}

func (t *TestTransport) GetBlockHash(height int64) (string, error) {
	return t.blocks[height].Hash, nil
}

func (t *TestTransport) GetBlockHeader(hash string) (*types.BlockHeader, error) {
	for _, header := range t.headers {
		if header.Hash == hash {
			return header, nil
		}
	}
	return nil, fmt.Errorf("block header not found: %s", hash)
}

func (t *TestTransport) GetBlockCount() (int64, error) {
	return t.blockCount, nil
}

func (t *TestTransport) GetBestBlockHash() (string, error) {
	return t.bestBlockHash, nil
}

func (t *TestTransport) GetBlockchainInfo() (*types.BlockchainInfo, error) {
	return t.blockChainInfo, nil
}

func (t *TestTransport) AddBlockAndHeader(block *types.Block, header *types.BlockHeader) error {
	t.blocks = append(t.blocks, block)
	t.headers = append(t.headers, header)
	return nil
}

func (t *TestTransport) SetBlockCount(count int64) error {
	t.blockCount = count
	return nil
}

func (t *TestTransport) SetBestBlockHash(hash string) error {
	t.bestBlockHash = hash
	return nil
}

func (t *TestTransport) SetBlockchainInfo(info *types.BlockchainInfo) error {
	t.blockChainInfo = info
	return nil
}

func NewTestTransport() *TestTransport {
	return &TestTransport{
		blocks:         []*types.Block{},
		headers:        []*types.BlockHeader{},
		bestBlockHash:  "",
		blockCount:     0,
		blockChainInfo: &types.BlockchainInfo{},
	}
}
