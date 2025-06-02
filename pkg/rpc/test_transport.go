package rpc

import (
	"fmt"

	"github.com/dogecoinfoundation/chainfollower/pkg/types"
)

type TestRpcTransport struct {
	RpcTransportInterface
	blocks         []*types.Block
	headers        []*types.BlockHeader
	bestBlockHash  string
	blockCount     int64
	blockChainInfo *types.BlockchainInfo
}

func (t *TestRpcTransport) GetBlock(hash string) (*types.Block, error) {
	for _, block := range t.blocks {
		if block.Hash == hash {
			return block, nil
		}
	}
	return nil, fmt.Errorf("block not found: %s", hash)
}

func (t *TestRpcTransport) GetBlockHash(height int64) (string, error) {
	return t.blocks[height].Hash, nil
}

func (t *TestRpcTransport) GetBlockHeader(hash string) (*types.BlockHeader, error) {
	for _, header := range t.headers {
		if header.Hash == hash {
			return header, nil
		}
	}
	return nil, fmt.Errorf("block header not found: %s", hash)
}

func (t *TestRpcTransport) GetBlockCount() (int64, error) {
	return t.blockCount, nil
}

func (t *TestRpcTransport) GetBestBlockHash() (string, error) {
	return t.bestBlockHash, nil
}

func (t *TestRpcTransport) GetBlockchainInfo() (*types.BlockchainInfo, error) {
	return t.blockChainInfo, nil
}

func (t *TestRpcTransport) AddBlockAndHeader(block *types.Block, header *types.BlockHeader) error {
	t.blocks = append(t.blocks, block)
	t.headers = append(t.headers, header)
	return nil
}

func (t *TestRpcTransport) SetBlockCount(count int64) error {
	t.blockCount = count
	return nil
}

func (t *TestRpcTransport) SetBestBlockHash(hash string) error {
	t.bestBlockHash = hash
	return nil
}

func (t *TestRpcTransport) SetBlockchainInfo(info *types.BlockchainInfo) error {
	t.blockChainInfo = info
	return nil
}

func NewTestRpcTransport() *TestRpcTransport {
	return &TestRpcTransport{
		blocks:         []*types.Block{},
		headers:        []*types.BlockHeader{},
		bestBlockHash:  "",
		blockCount:     0,
		blockChainInfo: &types.BlockchainInfo{},
	}
}
