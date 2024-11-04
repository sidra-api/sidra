package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// Connect to the server
	conn, err := net.Dial("unix", "/tmp/foo.sock")
	if err != nil {
		fmt.Println("Dial error:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Send a message to the server
	message := "Hello, Server!"
	_, err = conn.Write([]byte(message))
	if err != nil {
		fmt.Println("Write error:", err)
		return
	}

	// Receive a response from the server
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}

	fmt.Println("Server response:", string(buf[:n]))
}