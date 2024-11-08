package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"

	"github.com/sidra-gateway/sidra-plugins/lib"
)

func goPlugin(w http.ResponseWriter, r *http.Request) {
	pluginName := r.Header.Get("Plugins")
	if pluginName == "" {
		http.Error(w, "Missing Plugins header", http.StatusBadRequest)
		return
	}

	// Connect to the plugin's Unix socket
	socketPath := filepath.Join("/tmp", pluginName+".sock")
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		http.Error(w, "Plugin connection failed", http.StatusInternalServerError)
		log.Printf("Error connecting to plugin socket %s: %v\n", socketPath, err)
		return
	}
	defer conn.Close()

	// Marshal request data to JSON
	requestData := lib.SidraRequest{
		Headers: map[string]string{
			"X-Real-Ip": r.Header.Get("X-Real-Ip"),
			"Plugins":   pluginName,
		},
		Url:    r.URL.Path,
		Method: r.Method,
	}
	data, _ := json.Marshal(requestData)

	// Send the request to the plugin
	if _, err := conn.Write(data); err != nil {
		http.Error(w, "Failed to send data to plugin", http.StatusInternalServerError)
		log.Printf("Error writing to plugin socket: %v\n", err)
		return
	}

	// Read plugin response
	respData, err := io.ReadAll(conn)
	if err != nil {
		http.Error(w, "Failed to read plugin response", http.StatusInternalServerError)
		log.Printf("Error reading plugin response: %v\n", err)
		return
	}

	// Unmarshal response data
	var pluginResponse lib.SidraResponse
	if err := json.Unmarshal(respData, &pluginResponse); err != nil {
		http.Error(w, "Invalid plugin response format", http.StatusInternalServerError)
		log.Printf("Error unmarshalling plugin response: %v\n", err)
		return
	}

	// Write the response from the plugin back to the client
	w.WriteHeader(pluginResponse.StatusCode)
	w.Write([]byte(pluginResponse.Body))
	log.Printf("Response from plugin: %s\n", pluginResponse.Body)
}

func main() {
	http.HandleFunc("/", goPlugin)
	log.Println("Sidra plugin server running on port : 3080")
	if err := http.ListenAndServe(":3080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
