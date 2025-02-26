package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	// "sync"
)

var (
// totalCorrect   int
// totalIncorrect int
// mu             sync.Mutex
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		reader := bufio.NewReader(conn)
		buffer := make([]byte, 2048)

		n, err := reader.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Println("Client connection closed.")
				break
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("Connection timeout.")
				break
			}
			log.Println("Error reading from connection:", err)
			break
		}
		request := buffer[:n]

		response := ""

		if len(request) > 1 {
			fmt.Println("[INFO] Received request from Client")
			response = "Request processed successfully!"
		}

		if response != "" {
			_, err := conn.Write([]byte(response + "\n"))
			if err != nil {
				log.Println("Error sending response to client:", err)
			}
		}

	}
}

func main() {

	listener, err := net.Listen("tcp", ":3002")
	if err != nil {
		log.Fatal("Error starting TCP server:", err)
	}
	defer listener.Close()

	log.Println("Server started, waiting for connections...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}
