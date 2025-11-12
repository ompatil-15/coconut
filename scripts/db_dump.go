package main

import (
	"fmt"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"
)

func main() {
	dbPath := os.ExpandEnv("$HOME/.coconut/coconut.db")

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			fmt.Printf("Bucket: %s\n", name)

			return b.ForEach(func(k, v []byte) error {
				if v == nil {
					fmt.Printf("  Sub-bucket: %s\n", k)
				} else {
					fmt.Printf("  Key: %s -> %x\n", k, v)

				}
				return nil
			})
		})
	})
	return
}
