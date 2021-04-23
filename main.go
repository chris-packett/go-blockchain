package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"
)

type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

var Blockchain []Block

var mutex = &sync.Mutex{}

var bcServer chan []Block

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	bcServer = make(chan []Block)

	genesisBlock := Block{}
	genesisBlock = Block{
		Index:     0,
		Timestamp: time.Now().String(),
		BPM:       0,
		Hash:      calculateHash(genesisBlock),
		PrevHash:  "",
	}

	spew.Dump(genesisBlock)

	mutex.Lock()
	Blockchain = append(Blockchain, genesisBlock)
	mutex.Unlock()

	tcpPort := os.Getenv("PORT")

	server, err := net.Listen("tcp", ":"+tcpPort)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("TCP server listening on port:", tcpPort)

	defer server.Close()

	for {
		conn, err := server.Accept()

		if err != nil {
			log.Fatal(err)
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	io.WriteString(conn, "Enter a new BPM:")

	scanner := bufio.NewScanner(conn)

	go func() {
		for scanner.Scan() {
			bpm, err := strconv.Atoi(scanner.Text())

			if err != nil {
				log.Printf("%v is not a number: %v", scanner.Text(), err)
				continue
			}

			oldBlock := Blockchain[len(Blockchain)-1]

			newBlock, err := generateBlock(oldBlock, bpm)

			if err != nil {
				log.Println(err)
				continue
			}

			if isBlockValid(newBlock, oldBlock) {
				mutex.Lock()
				Blockchain = append(Blockchain, newBlock)
				mutex.Unlock()
			}

			bcServer <- Blockchain
			io.WriteString(conn, "\nEnter a new BPM:")
		}
	}()

	go func() {
		for {
			time.Sleep(30 * time.Second)

			mutex.Lock()
			output, err := json.MarshalIndent(Blockchain, "", "	")

			if err != nil {
				log.Fatal(err)
			}
			mutex.Unlock()

			io.WriteString(conn, "\n"+string(output))
			io.WriteString(conn, "\nEnter a new BPM:")
		}
	}()

	for range bcServer {
		spew.Dump(Blockchain)
	}
}

func isBlockValid(newBlock Block, oldBlock Block) bool {
	if newBlock.Index != oldBlock.Index+1 {
		return false
	}

	if newBlock.PrevHash != oldBlock.Hash {
		return false
	}

	if newBlock.Hash != calculateHash(newBlock) {
		return false
	}

	return true
}

func calculateHash(block Block) string {
	record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func generateBlock(oldBlock Block, bpm int) (Block, error) {
	var newBlock Block

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = time.Now().String()
	newBlock.BPM = bpm
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock, nil
}
