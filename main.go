package main

import (
	"fmt"
	"github.com/arhamj/simple_blockchain/blockchain"
	"strconv"
)

func main() {
	chain := blockchain.InitBlockchain()
	chain.AddBlock("Block #1")
	chain.AddBlock("Block #2")
	chain.AddBlock("Block #3")
	it := chain.Iterator()
	for {
		blk := it.Next()
		fmt.Printf("\tHash: %x\n", blk.Hash)
		fmt.Printf("\tPrevious Hash: %x\n", string(blk.PrevHash))
		fmt.Printf("\tData: %s\n", blk.Data)
		pow := blockchain.NewProof(blk)
		fmt.Printf("\tPoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		if blk.PrevHash == nil {
			break
		}
	}
}
