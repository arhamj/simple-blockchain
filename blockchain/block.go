package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash     []byte
	Txs      []*Tx
	PrevHash []byte
	Nonce    int
}

func (b *Block) HashTxs() []byte {
	var txHashes [][]byte
	var txHash [32]byte
	for _, tx := range b.Txs {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}

func CreateBlock(txs []*Tx, prevHash []byte) *Block {
	block := Block{
		PrevHash: prevHash,
		Txs:      txs,
		Nonce:    0,
	}
	pow := NewProof(&block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash
	return &block
}

func Genesis(coinbase *Tx) *Block {
	return CreateBlock([]*Tx{coinbase}, []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(b)
	HandleErr(err)
	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var res Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&res)
	HandleErr(err)
	return &res
}

func HandleErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}
