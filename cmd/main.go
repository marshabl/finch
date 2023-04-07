package main

import (
	"crypto/ecdsa"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const RESERVES_FILTER = "0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1"
const BUILDER_RPC_URL = "https://api.blocknative.com/v1/auction"

var ptt *PairsToTokens
var PRIVATEKEY string
var privateKey *ecdsa.PrivateKey
var myAddress common.Address
var rpc *RPC
var baseFee int64
var nonce uint64

func main() {
	APIKEY, ok := os.LookupEnv("APIKEY")
	if !ok {
		log.Printf("No environment variable APIRKEY.")
		return
	}

	RPCURL, ok := os.LookupEnv("RPCURL")
	if !ok {
		log.Printf("RPC url failed. Please try a different RPC url.")
		return
	}

	PRIVATEKEY, _ = os.LookupEnv("PRIVATEKEY")
	privateKey, _ = crypto.HexToECDSA(PRIVATEKEY)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("error casting public key to ECDSA")
	}
	myAddress = crypto.PubkeyToAddress(*publicKeyECDSA)

	ptt, _ = loadPairsToTokens()

	rpc = NewRPC(BUILDER_RPC_URL)

	go subscribeMempoolEvents(APIKEY)
	go subscribeLogEvents(RPCURL, ptt)
	go subscribeBlockEvents(RPCURL)
	for {
	}
}
