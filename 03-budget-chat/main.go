package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"sync"
)

var userRegexp = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

var broker = make(chan message)

var connections sync.Map

type connection struct {
	conn     net.Conn
	username *string
	messages chan string
}

type message struct {
	who  net.Conn
	data string
}

func roomContainsMessage(connections sync.Map) string {
	users := []string{}
	connections.Range(func(key, value interface{}) bool {
		username := value.(*connection).username
		if username != nil {
			users = append(users, *username)
		}
		return true
	})
	return "* The room contains: " + strings.Join(users, ", ") + "\n"
}

func joinMessage(conn connection) message {
	return message{
		who:  conn.conn,
		data: fmt.Sprintf("* %s has entered the room\n", *conn.username),
	}
}

func chatMessage(conn connection, chat string) message {
	return message{
		who:  conn.conn,
		data: fmt.Sprintf("[%s] %s\n", *conn.username, chat),
	}
}

func leaveMessage(conn connection) message {
	return message{
		who:  conn.conn,
		data: fmt.Sprintf("* %s has left the room\n", *conn.username),
	}
}

func main() {
	l, _ := net.Listen("tcp", ":9000")
	fmt.Printf("listening on :9000\n")
	defer l.Close()

	go func() {
		for {
			msg := <-broker
			connections.Range(func(key, value interface{}) bool {
				if key != msg.who && value.(*connection).username != nil {
					value.(*connection).messages <- msg.data
				}
				return true
			})
		}
	}()

	for {
		conn, _ := l.Accept()
		fmt.Printf("connection from %s\n", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	connection := connection{conn: conn, messages: make(chan string)}
	connections.Store(conn, &connection)

	reader := bufio.NewReader(conn)

	conn.Write([]byte("Welcome to the chat server! What should I call you?\n"))

	message, err := reader.ReadString('\n')
	if err != nil {
		if err != io.EOF {
			fmt.Printf("%s\n", err.Error())
		}
		goto disconnect
	}

	message = strings.TrimSpace(message)
	if !userRegexp.MatchString(message) {
		conn.Write([]byte("Invalid username. Disconnecting.\n"))
		goto disconnect
	}
	conn.Write([]byte(roomContainsMessage(connections)))
	connection.username = &message
	broker <- joinMessage(connection)

	go func() {
		for {
			message := <-connection.messages
			conn.Write([]byte(message))
		}
	}()

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				fmt.Printf("%s\n", err.Error())
			}
			goto disconnect
		}
		broker <- chatMessage(connection, strings.TrimSpace(message))
	}

disconnect:
	fmt.Printf("connection from %s closed\n", conn.RemoteAddr())
	conn.Close()
	connections.Delete(conn)
	if connection.username != nil {
		broker <- leaveMessage(connection)
	}
}
