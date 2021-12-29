package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := Block{
		PrevHash: prevHash,
		Data:     []byte(data),
		Nonce:    0,
	}
	pow := NewProof(&block)
	nonce, hash := pow.Run()
	block.Nonce = nonce
	block.Hash = hash
	return &block
}

func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
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
