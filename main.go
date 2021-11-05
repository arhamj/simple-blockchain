package main

import "github.com/boltdb/bolt"

func main() {
	bc := NewBlockchain()
	defer func(db *bolt.DB) {
		_ = db.Close()
	}(bc.db)

	cli := CLI{bc: bc}
	cli.Run()
}
