package main

import "sync"

type Blockchain struct {
	Blocks []*Block
	mu     sync.Mutex
}

func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks: []*Block{
			NewGenesisBlock(),
		},
	}
}

func (blockchain *Blockchain) AddBlock(newBlock *Block) {
	blockchain.mu.Lock()
	defer blockchain.mu.Unlock()
	blockchain.Blocks = append(blockchain.Blocks, newBlock)
}
