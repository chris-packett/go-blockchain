package main

import (
	"io"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
)

type Block struct {
	Timestamp string
	BPM       int
	PrevHash  string
	Hash      string
	Nonce     string
}

func NewBlock(w io.Writer, oldBlockHash string, bpm int) *Block {
	newBlock := &Block{
		Timestamp: time.Now().String(),
		BPM:       bpm,
		PrevHash:  oldBlockHash,
	}

	pow := NewProofOfWork(newBlock)
	nonce, hash := pow.Run(w)

	newBlock.Hash = hash
	newBlock.Nonce = nonce

	return newBlock
}

func (newBlock *Block) IsValid(oldBlockHash string) bool {
	if newBlock.PrevHash != oldBlockHash {
		return false
	}

	if newBlock.Hash != calculateHash(newBlock, newBlock.Nonce) {
		return false
	}
	return true
}

func NewGenesisBlock() *Block {
	genesisBlock := NewBlock(log.Writer(), "", 0)

	spew.Dump(genesisBlock)

	return genesisBlock
}
