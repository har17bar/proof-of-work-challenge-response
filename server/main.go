package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"reflect"
	"strconv"
	"time"
	"unsafe"
)

var texts = []string{"bbbbb", "DDDD"}
var chalangeDificulty uint8 = 3

var clients = make(map[string]int, 10000)

func main() {
	// Resolve the string address to a UDP address
	udpAddr, err := net.ResolveUDPAddr("udp", ":1200")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Start listening for UDP packages on the given address
	conn, err := net.ListenUDP("udp", udpAddr)
	defer conn.Close()

	if err != nil {
		fmt.Errorf("Error %s", err)
		return
	}

	// Read from UDP listener in endless loop
	for {
		var buf [256]byte
		_, addr, err := conn.ReadFromUDP(buf[0:])
		if err != nil {
			fmt.Errorf("Error %s", err)
			//return
		}

		fmt.Print("CLIENT MESSAGE: ", addr.IP, addr.Port)
		ip := addr.IP.String() + strconv.Itoa(addr.Port)

		attemptTime, exists := clients[ip]

		if exists {
			dif := time.Now().Second() - attemptTime
			fmt.Println(dif)
			if dif >= 5 {
				conn.WriteToUDP(authTokenRes(), addr)
			} else {
				data := &ReqBuff{}
				readMsgFromUdp(buf[0:], data)
				if isWorkDone(data) {
					fmt.Println("done")
					conn.WriteToUDP(authTokenRes(), addr)
				}
			}
		} else {
			clients[ip] = time.Now().Second()
		}

		seedId, text := getTextRand()
		answer := answer{Difficulty: chalangeDificulty, SeedId: seedId, Text: text, MaxTime: 5000}
		res, err := json.Marshal(answer)
		fmt.Println(string(res))
		if err != nil {
			fmt.Println(err)
		} else {
			res = append(res, '\n')
			// Write back the message over UPD
			conn.WriteToUDP(res, addr)
		}
	}
}

func getTextRand() (int64, string) {
	randNum := rand.Int63n((int64(len(texts))))
	return randNum, texts[randNum]
}

type answer struct {
	Text       string        `json:"text"`
	Difficulty uint8         `json:"difficulty"`
	SeedId     int64         `json:"seedId"`
	MaxTime    time.Duration `json:"maxTime"`
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Cap = sh.Len
	bh.Len = sh.Len
	return b
}

func isWorkDone(req *ReqBuff) bool {
	fmt.Println(string(req.Hash), "REQQQ HSAAA", req.SeedId, req.Nonce)
	hash := sha256.New()
	changedText := append(req.Hash, strconv.Itoa(int(req.Nonce))...)
	hash.Write(changedText)
	bs := hash.Sum(nil)
	hmacHex := []byte(hex.EncodeToString(bs))
	a := bytes.HasPrefix(hmacHex, bytes.Repeat([]byte{'0'}, int(chalangeDificulty)))
	fmt.Println(a, "HASPRIFAAAAA", string(changedText), string(hmacHex))
	if a {
		fmt.Println(a, "3635353----", string(hmacHex))
		return a
	}
	return a
}

func authTokenRes() []byte {
	return []byte("token")
}

type ReqBuff struct {
	Hash   []byte `json:"hash"`
	Nonce  uint64 `json:"nonce"`
	SeedId int64  `json:"seedId"`
}

func readMsgFromUdp(buf []byte, T interface{}) interface{} {
	// Find the null terminator index (assuming your message is null-terminated)
	nullIndex := -1
	for i, b := range buf {
		if b == 0 {
			nullIndex = i
			break
		}
	}

	if nullIndex == -1 {
		fmt.Println("Incomplete message received")
		//continue
	}

	// Extract the complete message
	message := buf[:nullIndex]

	reqParam := T
	err := json.Unmarshal(message, &reqParam)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		//continue
	}

	fmt.Println("Received message:", reqParam)
	return reqParam
}
