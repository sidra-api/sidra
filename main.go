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

//defaultHandler akan menangani request & memberikan response
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	serviceName := r.Header.Get("ServiceName")
	servicePort := r.Header.Get("ServicePort")
	plugins := r.Header.Get("Plugins")

	// Baca body awal
	bodyBytes, _ := io.ReadAll(r.Body)
	requestBody := string(bodyBytes)
	r.Body = io.NopCloser(strings.NewReader(requestBody)) // Reset body untuk keperluan penerusan berikutnya
	
	request := lib.SidraRequest{
		Headers: map[string]string{},
		Body:    requestBody,
		Url:     r.URL.String(),
		Method:  r.Method,
	}

	// Salin header
	for key, values := range r.Header {
		for _, value := range values {
			request.Headers[key] = value
		}
	}
	
	var response lib.SidraResponse

	// Jalankan plugin
	for _, plugin := range strings.Split(plugins, ",") {
		fmt.Println("execute " + plugin)
		response = goPlugin(plugin, request)

		//Set header dr plugin ke response
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}
		fmt.Printf("Response: %v", response.Headers["Cache-Control"])
		// Jika respons dr plugin tdk OK, kirimkan respons langsung ke client
		if response.StatusCode != http.StatusOK {
			fmt.Println("Plugin response not OK. Status: ", response.StatusCode)
			w.WriteHeader(response.StatusCode)
			w.Write([]byte(response.Body))
			return
		}
		if w.Header().Get("Cache-Control") != "no-cache" {
			w.WriteHeader(response.StatusCode)
			w.Write([]byte(response.Body))			
		}
	}

	if w.Header().Get("Cache-Control") == "no-cache" {
		err := forwardToService(w, r, serviceName, servicePort)
		if err != nil {
			http.Error(w, "Failed to forward request to service:"+err.Error(), http.StatusInternalServerError)
		}
	}
	for _, plugin := range strings.Split(plugins, ",") {
		fmt.Println("run plugin: " + plugin)
		response = goPlugin(plugin + ".response", request)
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

	requestBytes, _ := json.Marshal(request)
	conn.Write(requestBytes)
	buffer := make([]byte, 1024)
	n, _ := conn.Read(buffer[0:])
	json.Unmarshal(buffer[:n], &response)
	return response
}

func forwardToService(w http.ResponseWriter, r *http.Request, serviceName, servicePort string) error {
	targetURL := &url.URL{
		Scheme: "http",
		Host:   serviceName + ":" + servicePort,
		Path:   r.URL.Path,
	}

	// Re-read body to forward it to the service
	bodyBytes, _ := io.ReadAll(r.Body)
	
	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes))) // Reset body for future use in plugins

	proxyReq, _ := http.NewRequest(r.Method, targetURL.String(), io.NopCloser(strings.NewReader(string(bodyBytes))))

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
		return err
	}
	defer resp.Body.Close()

	// Write response headers and status code from target service to response writer
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	return nil
}

func main() {
	http.HandleFunc("/", defaultHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3080"
	}

	log.Println("Sidra plugin server is running on port :", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}