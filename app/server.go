package main

import (
	"fmt"
	"net"
	"os"
	"strings"
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

	fmt.Printf("Listening on %s\n", address)

	connection, err := listener.Accept()
	if err != nil {
		fmt.Printf("Error accepting connection: %v\n", err)
		os.Exit(1)
	}
	defer connection.Close()

	fmt.Printf("Client connected\n")

	readBuffer := make([]byte, 2048)
	bytesReceived, err := connection.Read(readBuffer)
	if err != nil {
		fmt.Printf("Error reading request: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Read %d bytes from client\n", bytesReceived)

	request := string(readBuffer[:bytesReceived])
	lines := strings.Split(request, "\r\n")
	parts := strings.Split(lines[0], " ")

	var statusCode int
	var statusMessage, path, responseBody string
	if len(parts) != 3 {
		statusCode, statusMessage = 500, "Bad Request"
	} else {
		path = parts[1]
		if path == "/" {
			statusCode, statusMessage = 200, "OK"
		} else if path == "/user-agent" {
			statusCode, statusMessage = 200, "OK"
			for _, line := range lines {
				if strings.HasPrefix(line, "User-Agent: ") {
					responseBody = line[12:]
				}
			}
		} else if strings.HasPrefix(path, "/echo/") {
			statusCode, statusMessage = 200, "OK"
			responseBody = path[6:]
		} else {
			statusCode, statusMessage = 404, "Not Found"
		}
	}

	httpResponse := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s\r\n",
		statusCode, statusMessage, len(responseBody), responseBody)

	bytesSent, err := connection.Write([]byte(httpResponse))
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sent %d bytes to client (expected: %d)\n", bytesSent, len(httpResponse))
}
