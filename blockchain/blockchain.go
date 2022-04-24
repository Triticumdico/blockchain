package blockchain

import (
	"fmt"
	"log"

	"github.com/dgraph-io/badger"
)

const dbPath = "./tmp/blocks"

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

// BlockChainIterator use to iterate through our blockachain which in database
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockChain() *BlockChain {
	var lastHash []byte

	opts := badger.DefaultOptions(dbPath)
	//opts.Dir = dbPath
	//opts.ValueDir = dbPath

	db, err := badger.Open(opts)
	if err != nil {
		log.Panic(err)
	}

	// Read and writte our transactions (txn) in our database
	err = db.Update(func(txn *badger.Txn) error {
		if _, err = txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing bockchain found")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serilize())
			if err != nil {
				log.Panic(err)
			}
			err = txn.Set([]byte("lh"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}

			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			if err != nil {
				log.Panic(err)
			}
			lastHash, err = item.ValueCopy(nil)
			return err
		}
	})
	if err != nil {
		log.Panic(err)
	}

	blockchain := BlockChain{lastHash, db}

	return &blockchain

}

func (chain *BlockChain) AddBlock(data string) {
	var lastHash []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			log.Panic(err)
		}

		lastHash, err = item.ValueCopy(nil)

		return err
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err = txn.Set(newBlock.Hash, newBlock.Serilize())
		if err != nil {
			log.Panic(err)
		}
		err = txn.Set([]byte("lh"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		chain.LastHash = newBlock.Hash
		return err
	})
	if err != nil {
		log.Panic(err)
	}
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedblock, err := item.ValueCopy(nil)
		block = Deserialize(encodedblock)

		return err
	})
	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}
