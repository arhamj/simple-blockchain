package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Tx struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value  int
	PubKey string
}

type TxInput struct {
	ID  []byte
	Out int
	Sig string
}

func NewTx(from, to string, amount int, chain *Chain) *Tx {
	var (
		inputs  []TxInput
		outputs []TxOutput
	)
	acc, validOutputs := chain.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("does not have enough funds")
	}

	for txIdStr, outs := range validOutputs {
		txId, err := hex.DecodeString(txIdStr)
		HandleErr(err)
		for _, out := range outs {
			input := TxInput{txId, out, from}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, TxOutput{amount, to})
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, to})
	}
	tx := Tx{
		Inputs:  inputs,
		Outputs: outputs,
	}
	tx.SetID()
	return &tx
}

func (t *Tx) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(t)
	HandleErr(err)
	hash = sha256.Sum256(encoded.Bytes())
	t.ID = hash[:]
}

func CoinbaseTx(to, data string) *Tx {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}
	txIn := TxInput{ID: []byte{}, Out: -1, Sig: data}
	txOut := TxOutput{Value: 100, PubKey: to}
	tx := Tx{
		ID:      nil,
		Inputs:  []TxInput{txIn},
		Outputs: []TxOutput{txOut},
	}
	tx.SetID()
	return &tx
}

// Validations

func (t *Tx) IsCoinbase() bool {
	return len(t.Inputs) == 1 && len(t.Inputs[0].ID) == 0 && t.Inputs[0].Out == -1
}

func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
