package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
)

const Difficulty = 12

type PoW struct {
	Block  *Block
	Target *big.Int
}

func NewProof(b *Block) *PoW {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))
	pow := &PoW{Block: b, Target: target}
	return pow
}

func (p *PoW) InitData(nonce int) []byte {
	data := bytes.Join([][]byte{
		p.Block.PrevHash,
		p.Block.HashTxs(),
		ToHex(int64(nonce)),
		ToHex(int64(Difficulty)),
	}, []byte{})
	return data
}

func (p *PoW) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte
	nonce := 0
	for nonce < math.MaxInt64 {
		data := p.InitData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])
		if intHash.Cmp(p.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println()
	return nonce, hash[:]
}

func (p *PoW) Validate() bool {
	var intHash big.Int
	data := p.InitData(p.Block.Nonce)
	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])
	return intHash.Cmp(p.Target) == -1
}

func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		panic(err)
	}
	return buff.Bytes()
}
