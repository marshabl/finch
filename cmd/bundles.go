package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// New create new rpc client with given url
func NewRPC(url string) *RPC {
	rpc := &RPC{
		url:     url,
		Headers: make(map[string]string),
		Timeout: 30 * time.Second,
	}
	rpc.client = &http.Client{
		Timeout: rpc.Timeout,
	}
	return rpc
}

func (rpc *RPC) newRPCRequest(method string, params ...interface{}) rpcRequest {
	return rpcRequest{
		ID:      1,
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
}

func (rpc *RPC) buildSignature(method string, privKey *ecdsa.PrivateKey, params ...interface{}) error {
	request := rpc.newRPCRequest(method, params...)

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	hashedBody := crypto.Keccak256Hash([]byte(body)).Hex()
	sig, err := crypto.Sign(accounts.TextHash([]byte(hashedBody)), privKey)
	if err != nil {
		return err
	}

	rpc.Signature = crypto.PubkeyToAddress(privKey.PublicKey).Hex() + ":" + hexutil.Encode(sig)
	rpc.RequestBody = body
	return nil
}

func (rpc *RPC) buildRPCRequest() error {
	req, err := http.NewRequest("POST", rpc.url, bytes.NewBuffer(rpc.RequestBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-Flashbots-Signature", rpc.Signature)
	for k, v := range rpc.Headers {
		req.Header.Add(k, v)
	}
	rpc.Request = req
	return nil
}

func (rpc *RPC) SendBundle(privKey *ecdsa.PrivateKey, txs []string, blockNumber string) (*SendBundleResponse, error) {
	opts := CallBundleParam{
		Txs:              txs,
		BlockNumber:      blockNumber,
		StateBlockNumber: "latest",
	}
	response, err := rpc.makeRpcCall("eth_sendBundle", privKey, opts)
	if err != nil {
		return nil, err
	}
	var sbResponse = new(SendBundleResponse)
	err = json.Unmarshal(response.Result, &sbResponse)
	return sbResponse, err
}

func (rpc *RPC) CallBundle(privKey *ecdsa.PrivateKey, txs []string, blockNumber string) (*CallBundleResponse, error) {
	opts := CallBundleParam{
		Txs:              txs,
		BlockNumber:      blockNumber,
		StateBlockNumber: "latest",
	}
	response, err := rpc.makeRpcCall("eth_callBundle", privKey, opts)
	if err != nil {
		return nil, err
	}

	var cbResponse = new(CallBundleResponse)
	err = json.Unmarshal(response.Result, &cbResponse)
	if err != nil {
		return nil, err
	}
	return cbResponse, nil
}

// CallWithFlashbotsSignature is like Call but also signs the request
func (rpc *RPC) makeRpcCall(method string, privKey *ecdsa.PrivateKey, params ...interface{}) (*rpcResponse, error) {
	if err := rpc.buildSignature(method, privKey, params...); err != nil {
		return nil, err
	}

	if err := rpc.buildRPCRequest(); err != nil {
		return nil, err
	}

	// fmt.Println(rpc)

	response, err := rpc.client.Do(rpc.Request)
	// fmt.Println(response)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	errorResp := new(RelayErrorResponse)
	if err := json.Unmarshal(data, errorResp); err == nil && errorResp.Error != "" {
		// relay returned an error
		return nil, fmt.Errorf("%w: %s", ErrRelayErrorResponse, errorResp.Error)
	}

	resp := new(rpcResponse)
	if err := json.Unmarshal(data, resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	return resp, nil
}
