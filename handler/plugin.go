package handler

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/sidra-api/sidra/dto"
)

func (h *Handler) GoPlugin(pluginName string, request dto.SidraRequest) (response dto.SidraResponse) {	
	//@TODO: for now body ignored
	request.Body = ""
	conn, err := net.Dial("unix", "/tmp/"+pluginName+".sock")
	if err != nil {
		return dto.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to connect to plugin: " + pluginName		,
		}
	}
	defer conn.Close()
	maxMessageSize := 16384
	if size, err := strconv.Atoi(os.Getenv("MAX_MESSAGE_SIZE")); err == nil {
		maxMessageSize = size
	} else {
		log.Printf("Invalid MAX_MESSAGE_SIZE value, using default: %v", err)
	}	
	requestBytes, _ := json.Marshal(request)
	conn.Write(requestBytes)
	buffer := make([]byte, maxMessageSize)
	n, _ := conn.Read(buffer[0:])
	json.Unmarshal(buffer[:n], &response)
	return response
}