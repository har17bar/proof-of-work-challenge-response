package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

func main() {

	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", ":1200")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Dial to the address with UDP
	conn, err := net.DialUDP("udp", nil, udpAddr)
	defer conn.Close()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Send a message to the server
	_, err = conn.Write([]byte("Client UDP Server1\n"))
	fmt.Println("send...")
	if err != nil {
		fmt.Errorf("Error %s", err)
		return
	}

	// Read from the connection until a new line is sent
	data, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		fmt.Errorf("Error %s", err)
		return
	}
	answer := Answer{}
	err = json.Unmarshal(data, &answer)
	if err != nil {
		fmt.Errorf("Error %s", err)
		return
	}
	hashCalculator := NewHashCalculator(answer.Text, answer.Difficulty)
	res := hashCalculator.solveChalange(time.Duration(answer.MaxTime) * time.Millisecond)
	res.SeedId = answer.SeedId
	json, err := json.Marshal(res)
	if err != nil {
		fmt.Errorf("Error %s", err)

	}
	// Print the data read from the connection to the terminal
	fmt.Println("Server MESSAGE: ", answer)

	fmt.Println(string(json), "++")
	fmt.Println(string(res.Hash), "22*()")
	conn.Write(json)
}

type Answer struct {
	Text       string `json:"text"`
	Difficulty uint8  `json:"difficulty"`
	SeedId     int64  `json:"seedId"`
	MaxTime    int    `json:"maxTime"`
}

type ResBuff struct {
	Hash   []byte `json:"hash"`
	Nonce  uint64 `json:"nonce"`
	SeedId int64  `json:"seedId"`
}

type HashCalculator struct {
	text       []byte
	difficulty []byte
}

func NewHashCalculator(text string, difficulty uint8) *HashCalculator {
	return &HashCalculator{text: []byte(text), difficulty: bytes.Repeat([]byte{'0'}, int(difficulty))}
}

func (h HashCalculator) solveChalange(t time.Duration) ResBuff {
	dateNow := time.Now()

	maxGorutine := uint64(runtime.NumCPU()) * 10 * uint64(len(h.difficulty))
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel() // Make sure it's called to release resources even if no errors

	fmt.Println("SolveChalange,", string(h.text), maxGorutine, uint64(runtime.NumCPU()))
	var i uint64 = 0
	var nonce uint64
	var hash []byte
	go func() {
		var wg sync.WaitGroup
		wg.Add(int(maxGorutine))
		for ; i < maxGorutine; i++ {
			go func(i uint64) {
				defer wg.Done()
				index, _hash := h.calculateHash(i)
				if _hash != nil {
					fmt.Println("FOUND:", i, string(_hash), index)
					nonce = index
					hash = _hash
					cancel()
				}
			}(i)
		}
		wg.Wait()
		fmt.Println("NOT FOUND")
		cancel()
	}()
	<-ctx.Done()
	fmt.Println("STOPPED")
	now := time.Now()
	fmt.Println("ELAPSED", dateNow.Sub(now.UTC()))
	fmt.Println(string(hash), nonce)
	return ResBuff{
		Nonce: nonce,
		Hash:  hash,
	}
}

func (h HashCalculator) calculateHash(startI uint64) (uint64, []byte) {
	var limit uint64 = 1000000
	var start uint64 = limit * startI
	var end uint64 = start + limit
	hash := sha256.New()
	for ; start <= end; start++ {
		changedText := h.text
		changedText = append(changedText, strconv.Itoa(int(start))...)
		hash.Write(changedText)
		bs := hash.Sum(nil)
		hmacHex := []byte(hex.EncodeToString(bs))
		if h.isValidHash(hmacHex) {
			return start, changedText
		}
	}
	return 0, nil
}

func (h HashCalculator) isValidHash(hash []byte) bool {
	a := bytes.HasPrefix(hash, h.difficulty)
	if a {
		fmt.Println(a, "&&&&&", string(hash))
		return a
	}
	return a
}
