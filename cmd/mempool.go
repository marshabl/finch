package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/marshabl/blocknative-go-sdk/bnsdk"
)

const SYSTEM = "ethereum"
const NETWORK = 1

const UNISWAP_UNIVERSAL_ROUTER = "0xEf1c6E67703c7BD7107eed8303Fbe6EC2554BF6B"
const UNISWAP_AUTO_ROUTER = "0xe592427a0aece92de3edee1f18e0157c05861564"
const UNISWAP_V3_ROUTER = "0x68b3465833fb72a70ecdf485e0e4c7bd8665fc45"
const UNISWAP_V2_ROUTER = "0x7a250d5630b4cf539739df2c5dacb4c659f2488d"
const SUSHISWAP_ROUTER = "0xd9e1ce17f2641f24ae83637ab66a2cca9c378b9f"
const ZEROX_EXCHANGE_PROXY = "0xdef1c0ded9bec7f1a1670819833240f027b25eff"
const ZEROX_COINBASE_PROXY = "0xe66b31678d6c16e9ebf358268a790b763c133750"
const ONE_INCH_V3 = "0x11111112542d85b3ef69ae05771c2dccff4faa26"
const ONE_INCH_V4 = "0x1111111254fb6c44bac0bed2854e76f90643097d"
const ONE_INCH_V5 = "0x1111111254eeb25477b68fb85ed929f73a960582"
const ONE_INCH_SETTLEMENT = "0xA88800CD213dA5Ae406ce248380802BD53b47647"
const PARASWAP_v4 = "0x1bd435f3c054b6e901b7b108a0ab7617c808677b"
const PARASWAP_v5 = "0xdef171fe48cf0115b1d80b88dc8eab59176fee57"
const COWSWAP_V2 = "0x9008d19f58aabd9ed0d60971565aa8510560ab41"
const METAMASK_SWAP_ROUTER = "0x881D40237659C251811CEC9c364ef91dC08D300C"
const SOCKET_REGISTRY = "0xc30141B657f4216252dc59Af2e7CdB9D8792e1B0"
const KYBER_MEGA_AGG_ROUTER = "0x6131b5fae19ea4f9d964eac0408e4408b66337b5"

const RETRYPERIOD = 10 * time.Second

func txnHandler(e []byte) {
	var tx = new(Payload)
	err := json.Unmarshal(e, tx)
	if err != nil {
		return
	}

	for _, balanceChange := range tx.Event.Transaction.NetBalanceChanges {
		pairAddress := balanceChange.Address

		if pair, ok := (*ptt)[pairAddress]; ok {
			// fmt.Println(tx.Event.Transaction.Hash, pairAddress, pair.PairInfo.Reserves0, pair.PairInfo.Reserves1)

			if len(pair.Cycles) > 0 {
				l := len(balanceChange.BalanceChanges) - 1
				adj0 := big.NewInt(0)
				adj1 := big.NewInt(0)
				token0 := common.HexToAddress(balanceChange.BalanceChanges[0].Asset.ContractAddress).String()
				token1 := common.HexToAddress(balanceChange.BalanceChanges[l].Asset.ContractAddress).String()

				adj0, ok = adj0.SetString(balanceChange.BalanceChanges[0].Delta, 10)
				if !ok {
					log.Printf("ERR", balanceChange)
					continue
				}

				adj1, ok = adj1.SetString(balanceChange.BalanceChanges[l].Delta, 10) // need error handling
				if !ok {
					log.Printf("ERR", balanceChange)
					continue
				}

				var changes = Adjustment{token0: adj0, token1: adj1}
				tr := getOptimalCycle(pair.Cycles, pairAddress, changes)
				fmt.Println("arb: ", tx.Event.Transaction.Hash, pairAddress, tr.optimalAmountIn, tr.maxProfit)

				if tr.maxProfit.Cmp(big.NewFloat(.002)) == 1 {
					gasPrice := big.NewInt(50000000000)
					nonce = uint64(17)
					data, _ := hex.DecodeString(tr.optimalData)
					myRawTxHash := getMyRawTxHash(nonce, gasPrice, data)

					t, err := BuildTxFromMpex(&tx.Event.Transaction)
					if err != nil {
						log.Printf("failed to build tx: %v", err)
					}

					rawTx := "0x" + TxToRlp(t)[6:]
					blockNumber := tx.Event.Transaction.PendingBlockNumber + 1
					blockNumberHex := IntToHex(blockNumber)

					txs := []string{rawTx, myRawTxHash}

					result, err := rpc.CallBundle(privateKey, txs, blockNumberHex)
					if err != nil {
						log.Println(err)
					}

					if result != nil {
						if result.Results[1].Error != "execution reverted" {
							fmt.Printf("%+v\n", result.Results)
							gasUsed := result.Results[1].GasUsed
							profit, _ := tr.maxProfit.Int64()
							margin := int64(95) / int64(100)
							gasPrice := profit * margin / gasUsed
							if gasPrice > baseFee {
								myNewRawTxHash := getMyRawTxHash(nonce, big.NewInt(gasPrice), data)
								txs = []string{txs[0], myNewRawTxHash}
								bundleResponse, err := rpc.SendBundle(privateKey, txs, blockNumberHex)
								if err != nil {
									log.Println(err)
								}

								fmt.Println(bundleResponse)
							} else {
								fmt.Println(gasPrice, baseFee)
							}
						}
					}
				}
			}
		}
	}
}

func errHandler(e error) {
	log.Printf("ERROR: %v", e)
}

func closeHandler(APIKEY string, sub *bnsdk.Subscription) {
	ticker := time.NewTicker(RETRYPERIOD)
	for {
		<-ticker.C
		log.Printf("Attempting to start a new connection...")
		err := startSubscription(APIKEY, sub)
		if err == nil {
			return
		}
	}
}

func startSubscription(APIKEY string, sub *bnsdk.Subscription) error {
	client, err := bnsdk.Stream(APIKEY, SYSTEM, NETWORK)
	if err != nil {
		log.Printf("error: %v", err.Error())
		return err
	}
	return client.Subscribe(sub)
}

func subscribeMempoolEvents(APIKEY string) {
	var filters []map[string]string
	filter := map[string]string{
		"status": "pending-simulation",
	}
	filters = append(filters, filter)
	config := bnsdk.NewConfig("global", filters)

	var monitor_addresses = []string{
		UNISWAP_UNIVERSAL_ROUTER,
		UNISWAP_AUTO_ROUTER,
		UNISWAP_V3_ROUTER,
		UNISWAP_V2_ROUTER,
		SUSHISWAP_ROUTER,
		ZEROX_EXCHANGE_PROXY,
		ZEROX_COINBASE_PROXY,
		ONE_INCH_V3,
		ONE_INCH_V4,
		ONE_INCH_V5,
		ONE_INCH_SETTLEMENT,
		PARASWAP_v4,
		PARASWAP_v5,
		COWSWAP_V2,
		METAMASK_SWAP_ROUTER,
		SOCKET_REGISTRY,
		KYBER_MEGA_AGG_ROUTER,
	}

	for _, address := range monitor_addresses {
		sub := bnsdk.NewSubscription(
			common.HexToAddress(address),
			txnHandler,
			errHandler,
			closeHandler,
			config,
		)
		go startSubscription(APIKEY, sub)
	}
}
