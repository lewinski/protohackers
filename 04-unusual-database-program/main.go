package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	l, _ := net.ListenPacket("udp", ":9000")
	fmt.Printf("listening on :9000\n")
	defer l.Close()

	store := make(map[string]string)

	for {
		buf := make([]byte, 1024)
		n, addr, err := l.ReadFrom(buf)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}

		req := string(buf[:n])
		fmt.Printf("%s: %s\n", addr, req)

		if req == "version" {
			l.WriteTo([]byte("version=lewinski's kv store 0.0"), addr)
			continue
		} else if strings.HasPrefix(req, "version=") {
			continue
		}

		if strings.Contains(req, "=") {
			t := strings.SplitN(req, "=", 2)
			store[t[0]] = t[1]
		} else {
			v, _ := store[req]
			l.WriteTo([]byte(fmt.Sprintf("%s=%s", req, v)), addr)
		}
	}
}
