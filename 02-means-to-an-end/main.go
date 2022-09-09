package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
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

	prices := map[int32]int32{}

loop:
	for {
		message := make([]byte, 9)
		n, err := io.ReadFull(reader, message)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("%s\n", err.Error())
			}
			break
		}
		if n != 9 {
			break
		}

		d1 := int32(binary.BigEndian.Uint32(message[1:5]))
		d2 := int32(binary.BigEndian.Uint32(message[5:9]))

		// fmt.Printf("%c %d %d\n", message[0], d1, d2)

		switch message[0] {
		case 'I':
			prices[d1] = int32(d2)
		case 'Q':
			tot := int64(0)
			cnt := int32(0)
			for tm, price := range prices {
				if tm >= d1 && tm <= d2 {
					tot += int64(price)
					cnt++
				}
			}
			avg := int32(0)
			if cnt > 0 {
				avg = int32(tot / int64(cnt))
			}
			binary.BigEndian.PutUint32(message[1:5], uint32(avg))
			conn.Write(message[1:5])
		default:
			fmt.Printf("invalid message from %s\n", conn.RemoteAddr())
			break loop
		}
	}

	fmt.Printf("connection from %s closed\n", conn.RemoteAddr())
	conn.Close()
}
