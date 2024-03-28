package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	var directory, host string
	var port int

	flag.StringVar(&host, "host", "0.0.0.0", "interface ip/host")
	flag.IntVar(&port, "port", 4221, "tcp port to listen for connections")
	flag.StringVar(&directory, "directory", ".", "directory from which to serve files")
	flag.Parse()

	info, err := os.Stat(directory)
	if err != nil {
		fmt.Printf("Failed to check directory path: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Invalid directory path %s\n", directory)
		os.Exit(1)
	}

	protocol := "tcp"
	address := fmt.Sprintf("%s:%d", host, port)

	listener, err := net.Listen(protocol, address)
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Listening for connections on %s\n", address)
	fmt.Printf("Serving files from %s\n", directory)

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
		}
		fmt.Printf("Client connected %v\n", connection.RemoteAddr())
		go handleConnection(connection, directory)
	}

}

func handleConnection(connection net.Conn, directory string) {

	defer connection.Close()

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
	var statusMessage, path, responseBody, requestFilePath string
	if len(parts) != 3 {
		statusCode, statusMessage = 400, "Bad Request"
	} else {
		statusCode, statusMessage = 200, "OK"
		path = parts[1]
		if path == "/" {
			// do nothing
		} else if path == "/user-agent" {
			statusCode, statusMessage = 200, "OK"
			for _, line := range lines {
				if strings.HasPrefix(line, "User-Agent: ") {
					responseBody = line[12:]
				}
			}
		} else if strings.HasPrefix(path, "/echo/") {
			responseBody = path[6:]
		} else if strings.HasPrefix(path, "/files/") {
			requestFilePath = path[7:]
		} else {
			statusCode, statusMessage = 404, "Not Found"
		}
	}

	if len(requestFilePath) > 0 {
		fullFilePath := filepath.Join(directory, requestFilePath)
		statusCode, statusMessage = handleFileRequest(connection, fullFilePath)
		if statusCode == 200 {
			return
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

func handleFileRequest(connection net.Conn, path string) (statusCode int, statusMessage string) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 404, "Not Found"
		}
		return 500, "Internal Server Error"
	}

	if info.IsDir() {
		return 500, "Internal Server Error"
	}

	file, err := os.Open(path)
	if err != nil {
		return 500, "Internal Server Error"
	}
	defer file.Close()

	size, _ := file.Seek(0, io.SeekEnd)
	statusCode, statusMessage = 200, "OK"
	httpHeader := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n",
		statusCode, statusMessage, size)
	_, err = connection.Write([]byte(httpHeader))
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
	}

	file.Seek(0, io.SeekStart)
	data := make([]byte, size)
	_, err = file.Read(data)
	if err != nil {
		return 500, "Internal Server Error"
	}

	_, err = connection.Write(data)
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
	}
	fmt.Printf("Served file %s to client\n", path)

	return
}
