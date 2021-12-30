package cli

import (
	"flag"
	"fmt"
	"github.com/arhamj/simple_blockchain/blockchain"
	"os"
	"runtime"
	"strconv"
)

type Cli struct{}

func (cli *Cli) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("\tgetBalance -address ADDRESS = get the balance for address")
	fmt.Println("\tcreateBlockchain -address ADDRESS = create a new blockchain")
	fmt.Println("\tprint = prints blocks in chain")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT = send amount")
}

func (cli *Cli) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *Cli) printChain() {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	it := chain.Iterator()
	for {
		blk := it.Next()
		fmt.Printf("\tHash: %x\n", blk.Hash)
		fmt.Printf("\tPrevious Hash: %x\n", string(blk.PrevHash))
		pow := blockchain.NewProof(blk)
		fmt.Printf("\tPoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
		if len(blk.PrevHash) == 0 {
			break
		}
	}
}

func (cli *Cli) createBlockchain(address string) {
	chain := blockchain.InitBlockchain(address)
	chain.Database.Close()
	fmt.Println("Created blockchain")
}

func (cli *Cli) getBalance(address string) {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	utxos := chain.FindUTXOs(address)
	balance := 0
	for _, out := range utxos {
		balance += out.Value
	}
	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *Cli) send(from, to string, amount int) {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	tx := blockchain.NewTx(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Tx{tx})
	fmt.Printf("Sent %d BTC from %s to %s", amount, from, to)
}

func (cli *Cli) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	getBalanceAddress := getBalanceCmd.String("address", "", "Address required")

	createBlockchainCmd := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	coinbaseAddress := createBlockchainCmd.String("address", "", "Address required")

	printBlockCmd := flag.NewFlagSet("print", flag.ExitOnError)

	sendTxCmd := flag.NewFlagSet("send", flag.ExitOnError)
	senderAddress := sendTxCmd.String("from", "", "Address required")
	receiverAddress := sendTxCmd.String("to", "", "Address required")
	sendAmount := sendTxCmd.Int("amount", 0, "Amount required")

	switch os.Args[1] {
	case "getBalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		blockchain.HandleErr(err)
	case "createBlockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		blockchain.HandleErr(err)
	case "print":
		err := printBlockCmd.Parse(os.Args[2:])
		blockchain.HandleErr(err)
	case "send":
		err := sendTxCmd.Parse(os.Args[2:])
		blockchain.HandleErr(err)
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *coinbaseAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockchain(*coinbaseAddress)
	}

	if printBlockCmd.Parsed() {
		cli.printChain()
	}

	if sendTxCmd.Parsed() {
		if *senderAddress == "" {
			sendTxCmd.Usage()
			runtime.Goexit()
		}
		if *receiverAddress == "" {
			sendTxCmd.Usage()
			runtime.Goexit()
		}
		if *sendAmount == 0 {
			sendTxCmd.Usage()
			runtime.Goexit()
		}
		cli.send(*senderAddress, *receiverAddress, *sendAmount)
	}
}
