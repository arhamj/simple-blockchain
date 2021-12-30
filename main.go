package main

import (
	"github.com/arhamj/simple_blockchain/cli"
	"os"
)

func main() {
	defer os.Exit(0)
	cmd := cli.Cli{}
	cmd.Run()
}
