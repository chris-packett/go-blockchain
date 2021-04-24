package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

var difficulty = 3

type ProofOfWork struct {
	Block *Block
}

func NewProofOfWork(block *Block) *ProofOfWork {
	return &ProofOfWork{
		Block: block,
	}
}

func (pow *ProofOfWork) Run(w io.Writer) (nonce string, hash string) {
	startTime := time.Now()

	for i := 0; ; i++ {
		nonce := fmt.Sprintf("%x", i)

		hash := calculateHash(pow.Block, nonce)

		if !isHashValid(hash, difficulty) {
			io.WriteString(w, hash+" do more work!\n")
			continue
		} else {
			io.WriteString(w, hash+" work done in "+time.Since(startTime).String()+"!\n")
			return nonce, hash
		}
	}
}

func calculateHash(block *Block, nonce string) string {
	record := block.Timestamp + strconv.Itoa(block.BPM) + block.PrevHash + nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}
