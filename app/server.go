package main

import (
	"fmt"
	"net"
	"os"
)

func main() {

	ip := "0.0.0.0"
	port := 4221
	protocol := "tcp"
	address := fmt.Sprintf("%s:%d", ip, port)

	listener, err := net.Listen(protocol, address)

	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
		os.Exit(1)
	}

	connection, err := listener.Accept()
	defer connection.Close()
	if err != nil {
		fmt.Printf("Error accepting connection: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Client connected\n")

	readBuffer := make([]byte, 2048)
	bytesReceived, err := connection.Read(readBuffer)
	if err != nil {
		fmt.Printf("Error reading request: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Read %d bytes from client\n", bytesReceived)

	httpResponse := "HTTP/1.1 200 OK\r\n\r\n"
	bytesSent, err := connection.Write([]byte(httpResponse))
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sent %d bytes to client (expected: %d)\n", bytesSent, len(httpResponse))
}
