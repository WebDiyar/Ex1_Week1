package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	CONN_PORT = ":3335"
	CONN_TYPE = "tcp"
)

var (
	clients     = make(map[net.Conn]string) // Changed to store nicknames
	clientsMux  sync.Mutex                  // Protects the clients map
	clientCount int                         // Track the number of connected clients
)

func broadcastMessage(message, nickname string, origin net.Conn) {
	logFile, err := os.OpenFile("history.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	messageWithNick := nickname + ": " + message // Prefix message with nickname
	timeStamp := time.Now().Format(time.RFC1123)
	logMessage := fmt.Sprintf("[%s] %s", timeStamp, messageWithNick)
	_, err = logFile.WriteString(logMessage + "\n") // Ensure newline at the end of each log entry
	if err != nil {
		log.Println("Failed to write to log file:", err)
	}

	clientsMux.Lock()
	for conn, _ := range clients {
		if conn != origin { // Don't send the message back to the sender
			conn.Write([]byte(messageWithNick))
		}
	}
	clientsMux.Unlock()
}

func handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		clientsMux.Lock()
		delete(clients, conn)
		clientCount--
		fmt.Printf("Client disconnected. Total clients: %d\n", clientCount)
		clientsMux.Unlock()
	}()

	// Prompt for nickname
	var nickname string
	fmt.Fprintln(conn, "Enter your nickname:")
	reader := bufio.NewReader(conn)
	nickname, err := reader.ReadString('\n')
	if err != nil {
		log.Println("Failed to read nickname:", err)
		return
	}
	nickname = strings.TrimSpace(nickname) // Remove newline character

	clientsMux.Lock()
	clients[conn] = nickname
	clientCount++
	fmt.Printf("Client connected: %s. Total clients: %d\n", nickname, clientCount)
	clientsMux.Unlock()

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Client disconnected")
			return
		}
		fmt.Print("Received from ", nickname, ": ", message)
		broadcastMessage(message, nickname, conn)
	}
}

func main() {
	listener, err := net.Listen(CONN_TYPE, CONN_PORT)
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	defer listener.Close()
	log.Println("Server is listening on " + CONN_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting: ", err)
			continue
		}
		go handleConnection(conn)
	}
}
