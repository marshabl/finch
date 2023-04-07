package main

import (
	"context"
	"encoding/hex"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func subscribeLogEvents(rpcUrl string, ptt *PairsToTokens) {
	client, err := ethclient.Dial(rpcUrl)
	if err != nil {
		log.Fatal(err)
	}

	var pairAddresses []common.Address

	for key, _ := range *ptt {
		address := common.HexToAddress(key)
		pairAddresses = append(pairAddresses, address)
	}

	topic := common.HexToHash(RESERVES_FILTER)
	topics := [][]common.Hash{{topic}}

	query := ethereum.FilterQuery{
		Addresses: pairAddresses,
		Topics:    topics,
	}

	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			pairAddress := vLog.Address.String()
			reserve0 := "0x" + strings.TrimLeft(hex.EncodeToString(vLog.Data[:32]), "0")
			reserve1 := "0x" + strings.TrimLeft(hex.EncodeToString(vLog.Data[32:]), "0")

			pttEntry := (*ptt)[pairAddress]
			pttEntry.PairInfo.Reserves0 = reserve0
			pttEntry.PairInfo.Reserves1 = reserve1

			(*ptt)[pairAddress] = pttEntry
			r0 := new(big.Int)
			r1 := new(big.Int)
			r0.SetString(pttEntry.PairInfo.Reserves0[2:], 16)
			r1.SetString(pttEntry.PairInfo.Reserves1[2:], 16)
			// fmt.Println("log: ", vLog.TxHash.String(), pttEntry.PairInfo.Token0, pttEntry.PairInfo.Token1, r0, r1)
		}
	}
}
