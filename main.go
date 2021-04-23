package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
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

type Message struct {
	BPM int
}

var mutex = &sync.Mutex{}

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal(err)
	}

	go func() {
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
	}()

	log.Fatal(run())
}

func run() error {
	mux := makeMuxRouter()
	httpPort := os.Getenv("PORT")

	log.Println("Listening on port", httpPort)

	s := &http.Server{
		Addr:           ":" + httpPort,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()

	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")

	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(Blockchain, "", "	")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	io.WriteString(w, string(bytes))
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}

	defer r.Body.Close()

	oldBlock := Blockchain[len(Blockchain)-1]

	newBlock, err := generateBlock(oldBlock, m.BPM)

	if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}

	if isBlockValid(newBlock, oldBlock) {
		Blockchain = append(Blockchain, newBlock)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, newBlock)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "	")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
		return
	}

	w.WriteHeader(code)
	w.Write(response)
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
