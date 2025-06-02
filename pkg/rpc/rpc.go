package rpc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"sync/atomic"

	"github.com/dogecoinfoundation/chainfollower/pkg/config"
	"github.com/dogecoinfoundation/chainfollower/pkg/types"
)

type rpcRequest struct {
	Method string `json:"method"`
	Params []any  `json:"params"`
	Id     uint64 `json:"id"`
}

type rpcResponse struct {
	Id     uint64           `json:"id"`
	Result *json.RawMessage `json:"result"`
	Error  any              `json:"error"`
}

type RpcTransport struct {
	RpcTransportInterface
	RpcClient *rpc.Client
	config    *config.Config
	Id        atomic.Uint64
}

func NewRpcTransport(config *config.Config) *RpcTransport {
	return &RpcTransport{config: config}
}

func (t *RpcTransport) GetBlock(hash string) (*types.Block, error) {
	res, err := t.Request("getblock", []any{hash, 2})
	if err != nil {
		return nil, err
	}

	var result *types.Block
	err = json.Unmarshal(*res, &result)
	if err != nil {
		return nil, fmt.Errorf("json-rpc unmarshal error: %v | %v", err, string(*res))
	}

	return result, nil
}

func (t *RpcTransport) GetBlockHash(height int64) (string, error) {
	res, err := t.Request("getblockhash", []any{height})
	if err != nil {
		return "", err
	}

	var result string
	err = json.Unmarshal(*res, &result)
	if err != nil {
		return "", fmt.Errorf("json-rpc unmarshal error: %v | %v", err, string(*res))
	}

	return result, nil
}

func (t *RpcTransport) GetBlockHeader(blockHash string) (header *types.BlockHeader, err error) {
	res, err := t.Request("getblockheader", []any{blockHash, true})
	if err != nil {
		return nil, err
	}

	var result *types.BlockHeader
	err = json.Unmarshal(*res, &result)
	if err != nil {
		return nil, fmt.Errorf("json-rpc unmarshal error: %v | %v", err, string(*res))
	}

	return result, nil
}

func (t *RpcTransport) GetBlockCount() (int64, error) {
	res, err := t.Request("getblockcount", []any{})
	if err != nil {
		return -1, err
	}

	var result int64
	err = json.Unmarshal(*res, &result)
	if err != nil {
		return -1, fmt.Errorf("json-rpc unmarshal error: %v | %v", err, string(*res))
	}

	return result, nil
}
func (t *RpcTransport) GetBestBlockHash() (string, error) {
	res, err := t.Request("getbestblockhash", []any{})
	if err != nil {
		return "", err
	}

	var result string
	err = json.Unmarshal(*res, &result)
	if err != nil {
		return "", fmt.Errorf("json-rpc unmarshal error: %v | %v", err, string(*res))
	}

	return result, nil
}

func (t *RpcTransport) GetBlockchainInfo() (*types.BlockchainInfo, error) {
	res, err := t.Request("getblockchaininfo", []any{})
	if err != nil {
		return nil, err
	}

	var result *types.BlockchainInfo
	err = json.Unmarshal(*res, &result)
	if err != nil {
		return nil, fmt.Errorf("json-rpc unmarshal error: %v | %v", err, string(*res))
	}

	return result, nil
}

func (t *RpcTransport) Request(method string, params []any) (*json.RawMessage, error) {
	id := t.Id.Add(1)

	body := rpcRequest{
		Method: method,
		Params: params,
		Id:     id,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("json-rpc marshal request: %v", err)
	}
	req, err := http.NewRequest("POST", t.config.RpcUrl, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("json-rpc request: %v", err)
	}

	if t.config.RpcUser != "" {
		req.SetBasicAuth(t.config.RpcUser, t.config.RpcPass)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("json-rpc transport: %v", err)
	}
	// we MUST read all of res.Body and call res.Close,
	// otherwise the underlying connection cannot be re-used.
	defer res.Body.Close()
	res_bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("json-rpc read response: %v", err)
	}
	// check for error response
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("json-rpc error status: %v | %v", res.StatusCode, string(res_bytes))
	}
	// cannot use json.NewDecoder: "The decoder introduces its own buffering
	// and may read data from r beyond the JSON values requested."
	var rpcres rpcResponse
	err = json.Unmarshal(res_bytes, &rpcres)
	if err != nil {
		return nil, fmt.Errorf("json-rpc unmarshal response: %v | %v", err, string(res_bytes))
	}
	if rpcres.Id != body.Id {
		return nil, fmt.Errorf("json-rpc wrong ID returned: %v vs %v", rpcres.Id, body.Id)
	}
	if rpcres.Error != nil {
		enc, err := json.Marshal(rpcres.Error)
		if err == nil {
			return nil, fmt.Errorf("json-rpc: error from Core Node: %v", string(enc))
		} else {
			return nil, fmt.Errorf("json-rpc: error from Core Node: %v", rpcres.Error)
		}
	}
	if rpcres.Result == nil {
		return nil, fmt.Errorf("json-rpc no result or error was returned")
	}

	return rpcres.Result, nil
}
