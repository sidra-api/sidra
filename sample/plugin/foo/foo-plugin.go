package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/sidra-gateway/sidra-plugins-hub/lib"
)

func main() {
	socketPath := "/tmp/foo.sock" // Tentukan lokasi socket

	// Pastikan socket tidak ada sebelum server berjalan
	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}

	// Buat listener pada UNIX socket
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		fmt.Println("Error creating listener:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on", socketPath)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Baca pesan dari client
	buffer := make([]byte, 32768)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading message:", err)
		return
	}
	message := buffer[:n]
	// Convert message to lib.SidraRequest
	var request lib.SidraRequest
	err = json.Unmarshal(message, &request)
	if err != nil {
		fmt.Println("Error unmarshalling message to SidraRequest:", err)
		return
	}
	fmt.Println(request)
	jsonStr, err := json.Marshal(request)
	if err != nil {
		fmt.Println("failed to marshal")
	}
	fmt.Println("Received message:", string(jsonStr))

	// Get headers from request and add some headers
	headers := request.Headers
	headers["X-Custom-Header"] = "CustomValue"
	response := lib.SidraResponse{
		StatusCode: http.StatusOK,
		Body:       "Success",
		Headers:    headers,
	}
	if request.Headers["Bar"] != "bar" {
		response = lib.SidraResponse{
			StatusCode: http.StatusForbidden,
			Body:       "Unauthorized",
			Headers:    headers,
		}
	}

	// Convert response to JSON
	responseBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshalling response:", err)
		return
	}
	fmt.Println("Response", string(responseBytes))
	conn.Write(responseBytes)

}
