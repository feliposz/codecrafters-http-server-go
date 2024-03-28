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

	_, err = listener.Accept()
	if err != nil {
		fmt.Printf("Error accepting connection: %v", err)
		os.Exit(1)
	}

}
