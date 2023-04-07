package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"time"
)

// most types pulled from https://github.com/metachris/flashbotsrpc/blob/master/types.go
type RpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RelayErrorResponse struct {
	Error string `json:"error"`
}

func (err RpcError) Error() string {
	return fmt.Sprintf("Error %d (%s)", err.Code, err.Message)
}

type CallBundleParam struct {
	Txs              []string `json:"txs"`                 // Array[String], A list of signed transactions to execute in an atomic bundle
	BlockNumber      string   `json:"blockNumber"`         // String, a hex encoded block number for which this bundle is valid on
	StateBlockNumber string   `json:"stateBlockNumber"`    // String, either a hex encoded number or a block tag for which state to base this simulation on. Can use "latest"
	Timestamp        int64    `json:"timestamp,omitempty"` // Number, the timestamp to use for this bundle simulation, in seconds since the unix epoch
	Timeout          int64    `json:"timeout,omitempty"`
	GasLimit         uint64   `json:"gasLimit,omitempty"`
	Difficulty       uint64   `json:"difficulty,omitempty"`
	BaseFee          uint64   `json:"baseFee,omitempty"`
}

type CallBundleResult struct {
	CoinbaseDiff      string `json:"coinbaseDiff"`      // "2717471092204423",
	EthSentToCoinbase string `json:"ethSentToCoinbase"` // "0",
	FromAddress       string `json:"fromAddress"`       // "0x37ff310ab11d1928BB70F37bC5E3cf62Df09a01c",
	GasFees           string `json:"gasFees"`           // "2717471092204423",
	GasPrice          string `json:"gasPrice"`          // "43000001459",
	GasUsed           int64  `json:"gasUsed"`           // 63197,
	ToAddress         string `json:"toAddress"`         // "0xdAC17F958D2ee523a2206206994597C13D831ec7",
	TxHash            string `json:"txHash"`            // "0xe2df005210bdc204a34ff03211606e5d8036740c686e9fe4e266ae91cf4d12df",
	Value             string `json:"value"`             // "0x"
	Error             string `json:"error"`
	Revert            string `json:"revert"`
}

type CallBundleResponse struct {
	BundleGasPrice    string             `json:"bundleGasPrice"`    // "43000001459",
	BundleHash        string             `json:"bundleHash"`        // "0x2ca9c4d2ba00d8144d8e396a4989374443cb20fb490d800f4f883ad4e1b32158",
	CoinbaseDiff      string             `json:"coinbaseDiff"`      // "2717471092204423",
	EthSentToCoinbase string             `json:"ethSentToCoinbase"` // "0",
	GasFees           string             `json:"gasFees"`           // "2717471092204423",
	Results           []CallBundleResult `json:"results"`           // [],
	StateBlockNumber  int64              `json:"stateBlockNumber"`  // 12960319,
	TotalGasUsed      int64              `json:"totalGasUsed"`      // 63197
}

type SendBundleRequest struct {
	Txs          []string  `json:"txs"`                         // Array[String], A list of signed transactions to execute in an atomic bundle
	BlockNumber  string    `json:"blockNumber"`                 // String, a hex encoded block number for which this bundle is valid on
	MinTimestamp *uint64   `json:"minTimestamp,omitempty"`      // (Optional) Number, the minimum timestamp for which this bundle is valid, in seconds since the unix epoch
	MaxTimestamp *uint64   `json:"maxTimestamp,omitempty"`      // (Optional) Number, the maximum timestamp for which this bundle is valid, in seconds since the unix epoch
	RevertingTxs *[]string `json:"revertingTxHashes,omitempty"` // (Optional) Array[String], A list of tx hashes that are allowed to revert
}

type SendBundleResponse struct {
	BundleHash string `json:"bundleHash"`
}

type rpcResponse struct {
	ID      int             `json:"id"`
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result"`
	Error   *RpcError       `json:"error"`
}

type rpcRequest struct {
	ID      int           `json:"id"`
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// FlashbotsRPC - Ethereum rpc client
type RPC struct {
	url         string
	client      httpClient
	Headers     map[string]string
	Signature   string
	RequestBody []byte
	Request     *http.Request
	Timeout     time.Duration
}

var ErrRelayErrorResponse = errors.New("relay error response")

type MpexTransaction struct {
	Nonce                *big.Int        `json:"nonce"`
	MaxPriorityFeePerGas string          `json:"maxPriorityFeePerGas,omitempty"`
	MaxFeePerGas         string          `json:"maxFeePerGas,omitempty"`
	Gas                  uint64          `json:"gas"`
	GasPrice             string          `json:"gasPrice,omitempty"`
	To                   string          `json:"to"`
	Value                string          `json:"value"`
	Input                string          `json:"input"`
	V                    string          `json:"v"`
	R                    string          `json:"r"`
	S                    string          `json:"s"`
	Type                 int             `json:"type"`
	Hash                 string          `json:"hash"`
	PendingBlockNumber   int             `json:"pendingBlockNumber"`
	NetBalanceChanges    []BalanceChange `json:"netBalanceChanges"`
}

type BalanceChange struct {
	Address        string
	BalanceChanges []Breakdown
}

type Breakdown struct {
	Delta string
	Asset Asset
}

type Asset struct {
	Type            string
	Symbol          string
	ContractAddress string
}

type Payload struct {
	Event struct {
		Transaction MpexTransaction
	}
}

type Adjustment map[string]*big.Int

type Reserves map[string]*big.Int

type TradeResult struct {
	maxProfit       *big.Float
	optimalCycle    []string
	optimalAmountIn *big.Int
	optimalSequence []Sequence
	optimalData     string
}

type Sequence struct {
	PairAddress string
	ReservesIn  *big.Int
	ReservesOut *big.Int
	Dir         string
}
