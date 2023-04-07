package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type PairsToTokens map[string]Pair

type Pair struct {
	PairInfo PairInfo   `json:"pairInfo"`
	Cycles   [][]string `json:"cycles"`
}

type PairInfo struct {
	Token0    string `json:"token0"`
	Token1    string `json:"token1"`
	Reserves0 string `json:"reserves0"`
	Reserves1 string `json:"reserves1"`
}

func loadPairsToTokens() (*PairsToTokens, error) {
	absPath, _ := filepath.Abs("../")

	path := filepath.Join(absPath, "graph/pairsToTokens.json")

	blob, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var ptt = new(PairsToTokens)
	err = json.Unmarshal(blob, ptt)
	if err != nil {
		return nil, err
	}

	// fmt.Println((*ptt)["0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc"].PairInfo.Reserves0)

	return ptt, nil
}
