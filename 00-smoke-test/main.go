package main

import (
	"fmt"
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
	buf := make([]byte, 1024)
	for {
		n, _ := conn.Read(buf)
		if n == 0 {
			break
		}

		conn.Write(buf[:n])
	}

	fmt.Printf("connection from %s closed\n", conn.RemoteAddr())
	conn.Close()
}
