package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
)

func main() {
	l, _ := net.Listen("tcp", ":9000")
	fmt.Printf("listening on :9000\n")
	defer l.Close()

	for {
		conn, _ := l.Accept()
		fmt.Printf("connection from %s\n", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("%s\n", err.Error())
			}
			break
		}
		// fmt.Printf("received: %s", message)

		req := struct {
			Method string   `json:"method"`
			Number *float64 `json:"number"`
		}{}

		err = json.Unmarshal([]byte(message), &req)
		if err != nil || req.Method != "isPrime" || req.Number == nil {
			conn.Write([]byte("{x\n"))
			continue
		}

		prime := false
		if math.Floor(*req.Number) == *req.Number {
			prime = big.NewInt(int64(*req.Number)).ProbablyPrime(0)
		}

		res := struct {
			Method string `json:"method"`
			Prime  bool   `json:"prime"`
		}{
			Method: "isPrime",
			Prime:  prime,
		}
		resBytes, _ := json.Marshal(res)
		conn.Write(append(resBytes, '\n'))
	}

	fmt.Printf("connection from %s closed\n", conn.RemoteAddr())
	conn.Close()
}
