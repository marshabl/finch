package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func getAmountOut(amountIn *big.Int, reservesIn *big.Int, reservesOut *big.Int) *big.Int {
	tFee := big.NewInt(997)
	bFee := big.NewInt(1000)
	var amountInWithFee = new(big.Int)
	var numerator = new(big.Int)
	var denominator = new(big.Int)
	var result = new(big.Int)

	amountInWithFee.Mul(amountIn, tFee)
	numerator.Mul(amountInWithFee, reservesOut)
	denominator.Add(reservesIn.Mul(reservesIn, bFee), amountInWithFee)

	return result.Div(numerator, denominator)
}

func getProfit(amountIn *big.Int, seqArray []Sequence) (*big.Float, string) {
	start := amountIn
	lenCycle := len(seqArray)
	lenCycleHex := fmt.Sprintf("%02x", lenCycle)
	amountInHex := fmt.Sprintf("%028x", amountIn)
	next := seqArray[0].PairAddress[2:]
	data := "00" + lenCycleHex + amountInHex + next
	for i, step := range seqArray {
		amountOut := getAmountOut(amountIn, step.ReservesIn, step.ReservesOut)
		amountIn = amountOut

		next = ""
		if i+1 < lenCycle {
			next = seqArray[i+1].PairAddress[2:]
		}

		dir := step.Dir
		amountOutHex := fmt.Sprintf("%028x", amountOut)
		data = data + next + amountOutHex + dir
	}
	var numerator = new(big.Int)
	var denominator = big.NewFloat(1000000000000000000) //ETH to WEI 10**18
	var result = new(big.Float)
	numerator.Sub(amountIn, start)
	n := new(big.Float).SetInt(numerator)
	result.Quo(n, denominator)
	return result, data
}

func getCycleReserves(cycle []string, pairAddress string, changes Adjustment) (Reserves, []Sequence) {
	var rs = make(Reserves)
	var seqArray = new([]Sequence)

	tokenIn := "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	tokenOut := "0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2"
	i := 1
	for _, address := range cycle {
		adj0 := big.NewInt(0)
		adj1 := big.NewInt(0)
		pairInfo := (*ptt)[address].PairInfo
		reserves0 := pairInfo.Reserves0
		reserves1 := pairInfo.Reserves1
		token0 := pairInfo.Token0
		token1 := pairInfo.Token1

		var reservesInHex string
		var reservesOutHex string
		if token0 == tokenOut {
			reservesInHex = reserves0
			reservesOutHex = reserves1
			tokenIn = token0
			tokenOut = token1
		} else {
			reservesInHex = reserves1
			reservesOutHex = reserves0
			tokenIn = token1
			tokenOut = token0
		}

		reservesIn := new(big.Int)
		reservesIn.SetString(reservesInHex[2:], 16)

		reservesOut := new(big.Int)
		reservesOut.SetString(reservesOutHex[2:], 16)

		if address == pairAddress {
			adj0 = changes[tokenIn]
			adj1 = changes[tokenOut]
		}

		r0 := big.NewInt(0)
		r1 := big.NewInt(0)
		if reservesIn != nil && reservesOut != nil && adj0 != nil && adj1 != nil {
			r0.Add(reservesIn, adj0)
			r1.Add(reservesOut, adj1)
		}

		key0 := "reserves" + strconv.Itoa(i) + "_" + strconv.Itoa(i+1)
		key1 := "reserves" + strconv.Itoa(i+1) + "_" + strconv.Itoa(i)
		rs[key0] = r0
		rs[key1] = r1

		dir := "00"
		if tokenIn < tokenOut {
			dir = "01"
		}

		step := new(Sequence)
		step.PairAddress = address
		step.Dir = dir
		step.ReservesIn = r0
		step.ReservesOut = r1
		*seqArray = append(*seqArray, *step)
		i++
	}

	return rs, *seqArray
}

func getOptimalCycle(cycles [][]string, pairAddress string, changes Adjustment) *TradeResult { //should return a new result type
	var tr = new(TradeResult)
	tr.maxProfit = big.NewFloat(0)
	tr.optimalAmountIn = big.NewInt(0)

	for _, cycle := range cycles {
		reserves, seqArray := getCycleReserves(cycle, pairAddress, changes)
		numPairs := len(cycle)
		amountIn := big.NewFloat(0)
		one := big.NewFloat(1)
		two := big.NewFloat(1)
		denominator := big.NewFloat(0)

		r1Num := new(big.Float)
		r2Num := new(big.Float)
		r1r2Num := new(big.Float)
		r1Den := new(big.Float)
		r2Den := new(big.Float)

		for i := 1; i < numPairs+1; i++ {
			r1Num.SetInt(reserves["reserves"+strconv.Itoa(i)+"_"+strconv.Itoa(i+1)])
			r2Num.SetInt(reserves["reserves"+strconv.Itoa(i+1)+"_"+strconv.Itoa(i)])

			r1r2Num.Mul(r1Num, r2Num)
			one.Mul(one, r1r2Num)
			two.Mul(two, r1Num)

			three := big.NewFloat(1)
			four := big.NewFloat(1)

			for j := 1; j < i; j++ {
				r1Den.SetInt(reserves["reserves"+strconv.Itoa(j+1)+"_"+strconv.Itoa(j)])
				three.Mul(three, r1Den)
			}

			for j := i + 1; j < numPairs+1; j++ {
				r2Den.SetInt(reserves["reserves"+strconv.Itoa(j)+"_"+strconv.Itoa(j+1)])
				four.Mul(four, r2Den)
			}

			insideFee := big.NewFloat(math.Pow(0.997, float64(i)))
			three.Mul(three, four)
			three.Mul(three, insideFee)
			denominator.Add(denominator, three)
		}

		outsideFee := big.NewFloat(math.Pow(0.997, float64(numPairs)))
		one.Mul(one, outsideFee)
		one.Sqrt(one)
		numerator := big.NewFloat(0)
		numerator.Sub(one, two)

		amountIn.Quo(numerator, denominator)

		if amountIn.Cmp(big.NewFloat(0)) == 1 {
			amountInInt64, _ := amountIn.Int64()
			amountInBigInt := big.NewInt(amountInInt64)
			profit, data := getProfit(amountInBigInt, seqArray)
			if profit.Cmp(tr.maxProfit) == 1 {
				tr.maxProfit = profit
				tr.optimalCycle = cycle
				tr.optimalAmountIn = amountInBigInt
				tr.optimalSequence = seqArray
				tr.optimalData = data
			}
		}
	}
	return tr
}

func convertStringToBigInt(s string, base int) (*big.Int, error) {
	ret := new(big.Int)
	ret, ok := ret.SetString(s, base)
	if !ok {
		return nil, fmt.Errorf("failed to set string %s to base %v", s, base)
	}
	return ret, nil
}

func IntToHex(i int) string {
	return fmt.Sprintf("0x%x", i)
}

func TxToRlp(tx *types.Transaction) string {
	var buff bytes.Buffer
	tx.EncodeRLP(&buff)
	return fmt.Sprintf("%x", buff.Bytes())
}

func BuildTxFromMpex(m *MpexTransaction) (*types.Transaction, error) {
	chainId := big.NewInt(1)

	nonce := m.Nonce.Uint64()

	gas := m.Gas

	to := common.HexToAddress(m.To)

	value, err := convertStringToBigInt(m.Value, 10)
	if err != nil {
		return nil, err
	}

	data, err := hex.DecodeString(m.Input[2:])
	if err != nil {
		return nil, err
	}

	v, err := convertStringToBigInt(m.V[2:], 16)
	if err != nil {
		return nil, err
	}

	r, err := convertStringToBigInt(m.R[2:], 16)
	if err != nil {
		return nil, err
	}

	s, err := convertStringToBigInt(m.S[2:], 16)
	if err != nil {
		return nil, err
	}

	if m.Type == 2 {
		gasTipCap, err := convertStringToBigInt(m.MaxPriorityFeePerGas, 10)
		if err != nil {
			return nil, err
		}

		gasFeeCap, err := convertStringToBigInt(m.MaxFeePerGas, 10)
		if err != nil {
			return nil, err
		}

		dynamicTx := types.DynamicFeeTx{
			ChainID:   chainId,
			Nonce:     nonce,
			GasTipCap: gasTipCap,
			GasFeeCap: gasFeeCap,
			Gas:       gas,
			To:        &to,
			Value:     value,
			Data:      data,
			V:         v,
			R:         r,
			S:         s,
		}
		return types.NewTx(&dynamicTx), nil
	}

	if m.Type == 0 {
		gasPrice, err := convertStringToBigInt(m.GasPrice, 10)
		if err != nil {
			return nil, err
		}

		legacyTx := types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gas,
			To:       &to,
			Value:    value,
			Data:     data,
			V:        v,
			R:        r,
			S:        s,
		}
		return types.NewTx(&legacyTx), nil
	}

	return nil, fmt.Errorf("unable to convert MPEX transaction to GETH transaction type")

}

func getMyRawTxHash(nonce uint64, gasPrice *big.Int, data []byte) string {
	chainId := big.NewInt(1)
	value := big.NewInt(0)
	gasLimit := uint64(250000)

	toAddress := common.HexToAddress("0x6d3797426B1CCf0cF5028Cd7A27d6b100e5D378b")
	legacyTx := types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    value,
		Data:     data,
	}
	tx := types.NewTx(&legacyTx)

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
	if err != nil {
		log.Fatal(err)
	}

	return "0x" + TxToRlp(signedTx)
}
