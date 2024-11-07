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
	"strconv"
	"strings"

	"github.com/sidra-gateway/sidra-plugins/lib"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {

	serviceName := r.Header.Get("ServiceName")
	servicePort := r.Header.Get("ServicePort")
	plugins := r.Header.Get("Plugins")
	
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
		Method: r.Method,
	}

	for key, values := range r.Header {
		for _, value := range values {
			request.Headers[key] = value
		}
	}
	response := lib.SidraResponse {
		StatusCode: http.StatusOK,
		Headers:    make(map[string]string),
		Body:       "",
	}
	for _, plugin := range strings.Split(plugins, ",") {
		fmt.Println("execute " + plugin)
		response = goPlugin(plugin, request)
		fmt.Println("response " + strconv.Itoa(response.StatusCode))
		if response.StatusCode != http.StatusOK {
			break;
		}
	}
	if response.StatusCode != http.StatusOK {
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(response.StatusCode)
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
	conn, err := net.Dial("unix", "/tmp/" + pluginName + ".sock")
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

	n, err := conn.Read(buffer)
	responseBytes := buffer[:n]
	if err != nil {
		return lib.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to read response from plugin: " + err.Error(),
		}
	}
	
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

	proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		return err
	}

	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	return err
}

func main() {
	http.HandleFunc("/", defaultHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Sidra plugin server running on port :", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}