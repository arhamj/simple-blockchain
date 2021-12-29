package blockchain

import (
	"fmt"
	"github.com/dgraph-io/badger"
)

const (
	dbPath = "./tmp/blocks"
	lhKey  = "lh"
)

type Chain struct {
	LashHash []byte
	Database *badger.DB
}

type ChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockchain() *Chain {
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	HandleErr(err)
	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get([]byte(lhKey)); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			HandleErr(err)
			err = txn.Set([]byte(lhKey), genesis.Hash)
			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte(lhKey))
			HandleErr(err)
			lastHash, err = item.ValueCopy(lastHash)
			return err
		}
	})
	HandleErr(err)
	return &Chain{Database: db, LashHash: lastHash}
}

func (c *Chain) AddBlock(data string) {
	var lastHash []byte
	err := c.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lhKey))
		HandleErr(err)
		lastHash, err = item.ValueCopy(lastHash)
		return err
	})
	HandleErr(err)
	newBlock := CreateBlock(data, lastHash)
	err = c.Database.Update(func(txn *badger.Txn) error {
		err = txn.Set(newBlock.Hash, newBlock.Serialize())
		HandleErr(err)
		err = txn.Set([]byte(lhKey), newBlock.Hash)
		lastHash = newBlock.Hash
		return err
	})
	c.LashHash = lastHash
}

func (c *Chain) Iterator() *ChainIterator {
	return &ChainIterator{
		CurrentHash: c.LashHash,
		Database:    c.Database,
	}
}

func (i *ChainIterator) Next() *Block {
	var block *Block
	err := i.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(i.CurrentHash)
		HandleErr(err)
		var encodedBlock []byte
		encodedBlock, err = item.ValueCopy(encodedBlock)
		block = Deserialize(encodedBlock)
		return err
	})
	HandleErr(err)
	i.CurrentHash = block.PrevHash
	return block
}
