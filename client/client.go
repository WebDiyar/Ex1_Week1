package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	CONN_PORT = ":3335"
	CONN_TYPE = "tcp"
)

var wg sync.WaitGroup

func read(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Disconnected from server.")
			wg.Done()
			return
		}
		fmt.Print("Server says: ", message)
	}
}

func write(conn net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(conn)

	for {
		fmt.Print("Enter message: ")
		message, _ := reader.ReadString('\n')
		_, err := writer.WriteString(message)
		if err != nil {
			fmt.Println("Error writing to server:", err)
			return
		}
		writer.Flush()
	}
}

func main() {
	fmt.Println("Enter 'join' to connect to the chat server.")
	reader := bufio.NewReader(os.Stdin)
	command, _ := reader.ReadString('\n')

	if strings.TrimSpace(command) == "join" {
		conn, err := net.Dial(CONN_TYPE, CONN_PORT)
		if err != nil {
			fmt.Println("Error connecting to server:", err)
			return
		}
		defer conn.Close()

		wg.Add(2)
		go func() {
			defer wg.Done()
			read(conn)
		}()
		go func() {
			defer wg.Done()
			write(conn)
		}()
		wg.Wait()
	} else {
		fmt.Println("Unknown command. Exiting.")
	}
}
