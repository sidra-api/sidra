package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/sidra-gateway/sidra-plugins/lib"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	serviceName := r.Header.Get("ServiceName")
	servicePort := r.Header.Get("ServicePort")
	plugins := r.Header.Get("Plugins")

	// Baca body awal
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body.", http.StatusInternalServerError)
		return
	}
	requestBodyString := string(bodyBytes)
	request := lib.SidraRequest{
		Headers: make(map[string]string),
		Body:    requestBodyString,
		Url:     r.URL.String(),
		Method:  r.Method,
	}

	// Reset body untuk keperluan penerusan berikutnya
	r.Body = io.NopCloser(strings.NewReader(requestBodyString))

	// Salin header
	for key, values := range r.Header {
		for _, value := range values {
			request.Headers[key] = value
		}
	}

	response := lib.SidraResponse{
		StatusCode: http.StatusOK,
		Headers:    make(map[string]string),
		Body:       "",
	}

	// Jalankan plugin
	for _, plugin := range strings.Split(plugins, ",") {
		fmt.Println("execute " + plugin)
		response = goPlugin(plugin, request)
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}
		
		// Jika respons dr plugin tdk OK, kirimkan respons langsung ke client
	}
	if response.StatusCode != http.StatusOK {
		w.WriteHeader(response.StatusCode)
		w.Write([]byte(response.Body))
		return
	}

	// TODO : Bikin plugin cache by URL pakai Redis
	if response.Headers["x-cache"] == "1" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response.Body))
		return
	}
	if serviceName == "" || servicePort == "" {
		http.Error(w, "ServiceName or ServicePort are not available.", http.StatusBadRequest)
		return
	}

	err = forwardToService(w, r, serviceName, servicePort)
	if err != nil {
		log.Println("Failed to forward request :", err)
		http.Error(w, "Failed to forward request.", http.StatusInternalServerError)
	}
}

func goPlugin(pluginName string, request lib.SidraRequest) (response lib.SidraResponse) {
	conn, err := net.Dial("unix", "/tmp/"+pluginName+".sock")
	if err != nil {
		return lib.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to connect to plugin: " + err.Error(),
		}
	}
	defer conn.Close()

	requestBytes, err := json.Marshal(request)
	if err != nil {
		return lib.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to marshal request: " + err.Error(),
		}
	}
	fmt.Println("write to plugins")
	_, err = conn.Write(requestBytes)
	if err != nil {
		return lib.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to send request to plugin: " + err.Error(),
		}
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer[0:])
	if err != nil {
		return lib.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to read response from plugin: " + err.Error(),
		}
	}
	responseBytes := buffer[:n]
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return lib.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to unmarshal response: " + err.Error(),
		}
	}
	return
}

func forwardToService(w http.ResponseWriter, r *http.Request, serviceName, servicePort string) error {
	targetURL := &url.URL{
		Scheme: "http",
		Host:   serviceName + ":" + servicePort,
		Path:   r.URL.Path,
	}

	// Re-read body to forward it to the service
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading request body: %v", err)
	}
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes))) // Reset body for future use in plugins

	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), io.NopCloser(strings.NewReader(string(bodyBytes))))
	if err != nil {
		return fmt.Errorf("error creating proxy request: %v", err)
	}

	// Copy headers from original request to new request
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Send the request to the target service
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("error forwarding to target service: %v", err)
	}
	defer resp.Body.Close()

	// Write response headers and status code from target service to response writer
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// Copy the response body from target service to the response writer
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return fmt.Errorf("error copying response body: %v", err)
	}

	return nil
}

func main() {
	http.HandleFunc("/", defaultHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3080"
	}

	log.Println("Sidra plugin server running on port :", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
