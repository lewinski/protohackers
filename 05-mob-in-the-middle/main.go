package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
)

type ProxyConnection struct {
	client     net.Conn
	clientData chan string
	server     net.Conn
	serverData chan string
}

func (proxy *ProxyConnection) Close() {
	proxy.client.Close()
	proxy.server.Close()
}

func (proxy *ProxyConnection) receiveFromClient() {
	reader := bufio.NewReader(proxy.client)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("%s\n", err.Error())
			}
			break
		}
		proxy.clientData <- strings.TrimSpace(message)
	}
	close(proxy.clientData)
}

func (proxy *ProxyConnection) receiveFromServer() {
	reader := bufio.NewReader(proxy.server)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("%s\n", err.Error())
			}
			break
		}
		proxy.serverData <- strings.TrimSpace(message)
	}
	close(proxy.serverData)
}

func makeProxyConnection(client net.Conn) *ProxyConnection {
	upstream, err := net.Dial("tcp", "chat.protohackers.com:16963")
	if err != nil {
		fmt.Println("error connecting to upstream", err)
		panic(err)
	}

	proxy := ProxyConnection{
		client:     client,
		clientData: make(chan string),
		server:     upstream,
		serverData: make(chan string),
	}

	go proxy.receiveFromClient()
	go proxy.receiveFromServer()

	return &proxy
}

func main() {
	l, _ := net.Listen("tcp", ":9000")
	fmt.Printf("listening on :9000\n")
	defer l.Close()

	for {
		conn, _ := l.Accept()
		fmt.Printf("connection from %s\n", conn.RemoteAddr())

		proxy := makeProxyConnection(conn)
		go handleConnection(proxy)
	}
}

func hack(message string) string {
	bogusRx := regexp.MustCompile("^7[0-9a-zA-Z]{25,34}$")
	tony := "7YWHMfk9JZe0LM0g1ZauHuiSxhI"

	s := strings.Split(message, " ")
	for i, word := range s {
		if bogusRx.MatchString(word) {
			s[i] = tony
		}
	}

	return strings.Join(s, " ")
}

func handleConnection(proxy *ProxyConnection) {
	for {
		select {
		case clientData := <-proxy.clientData:
			if clientData == "" {
				goto disconnect
			}
			fmt.Printf("client: %s\n", clientData)
			proxy.server.Write([]byte(hack(clientData) + "\n"))
		case serverData := <-proxy.serverData:
			if serverData == "" {
				goto disconnect
			}
			fmt.Printf("server: %s\n", serverData)
			proxy.client.Write([]byte(hack(serverData) + "\n"))
		}
	}
disconnect:
	proxy.Close()
}
