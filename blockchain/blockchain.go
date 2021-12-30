package blockchain

import (
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"os"
	"runtime"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First tx from genesis"
	lhKey       = "lh"
)

type Chain struct {
	LashHash []byte
	Database *badger.DB
}

type ChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockchain(address string) *Chain {
	var lastHash []byte
	if DBExists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	HandleErr(err)
	err = db.Update(func(txn *badger.Txn) error {
		cbTx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbTx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		HandleErr(err)
		err = txn.Set([]byte(lhKey), genesis.Hash)
		lastHash = genesis.Hash
		return err
	})
	HandleErr(err)
	return &Chain{Database: db, LashHash: lastHash}
}

func ContinueBlockchain(address string) *Chain {
	var lastHash []byte
	if !DBExists() {
		fmt.Println("Blockchain does not exist, create one!")
		runtime.Goexit()
	}
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	HandleErr(err)
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lhKey))
		HandleErr(err)
		lastHash, err = item.ValueCopy(lastHash)
		return err
	})
	HandleErr(err)
	return &Chain{Database: db, LashHash: lastHash}
}

func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func (c *Chain) AddBlock(txs []*Tx) {
	var lastHash []byte
	err := c.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(lhKey))
		HandleErr(err)
		lastHash, err = item.ValueCopy(lastHash)
		return err
	})
	HandleErr(err)
	newBlock := CreateBlock(txs, lastHash)
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

func (c *Chain) FindUTXs(address string) []Tx {
	var utxs []Tx
	spentTXOs := make(map[string][]int)
	iter := c.Iterator()
	for {
		block := iter.Next()
		for _, tx := range block.Txs {
			txId := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txId] != nil {
					for _, spentOut := range spentTXOs[txId] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlocked(address) {
					utxs = append(utxs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.CanUnlock(address) {
						inTxId := hex.EncodeToString(in.ID)
						spentTXOs[inTxId] = append(spentTXOs[inTxId], in.Out)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return utxs
}

func (c *Chain) FindUTXOs(address string) []TxOutput {
	var utxos []TxOutput
	utxs := c.FindUTXs(address)
	for _, tx := range utxs {
		for _, txOut := range tx.Outputs {
			if txOut.CanBeUnlocked(address) {
				utxos = append(utxos, txOut)
			}
		}
	}
	return utxos
}

func (c *Chain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	utxos := make(map[string][]int)
	utxs := c.FindUTXs(address)
	accumulated := 0

Work:
	for _, tx := range utxs {
		txId := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.Outputs {
			if out.CanBeUnlocked(address) && accumulated < amount {
				accumulated += out.Value
				utxos[txId] = append(utxos[txId], outIdx)
				if accumulated >= amount {
					break Work
				}
			}
		}
	}
	return accumulated, utxos
}
