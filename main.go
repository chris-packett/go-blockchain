package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"
	env "github.com/joho/godotenv"
)

var (
	blockchain       *Blockchain
	blockchainServer chan *Blockchain
)

func main() {
	if err := env.Load(); err != nil {
		log.Fatal(err)
	}

	blockchainServer = make(chan *Blockchain)

	blockchain = NewBlockchain()

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

	connLogger := log.New(conn, "", log.LstdFlags)

	connLogger.Print("Enter a new BPM:")

	scanner := bufio.NewScanner(conn)

	go func() {
		for scanner.Scan() {
			bpm, err := strconv.Atoi(scanner.Text())

			if err != nil {
				connLogger.Printf("%v is not a number: %v", scanner.Text(), err)
				continue
			}

			oldBlock := blockchain.Blocks[len(blockchain.Blocks)-1]
			newBlock := NewBlock(conn, oldBlock.Hash, bpm)

			if newBlock.IsValid(oldBlock.Hash) {
				blockchain.AddBlock(newBlock)
			}

			blockchainServer <- blockchain
			connLogger.Print("Enter a new BPM:")
		}
	}()

	go func() {
		for {
			time.Sleep(30 * time.Second)

			output, err := json.MarshalIndent(blockchain, "", "	")

			if err != nil {
				log.Fatal(err)
			}

			connLogger.Print(string(output))
			connLogger.Print("Enter a new BPM:")
		}
	}()

	for range blockchainServer {
		spew.Dump(blockchain.Blocks)
	}
}
