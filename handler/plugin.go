package handler

import (
	"encoding/json"	
	"net"
	"net/http"

	"github.com/sidra-api/sidra/dto"
)

func (h *Handler) GoPlugin(pluginName string, request dto.SidraRequest) (response dto.SidraResponse) {	
	conn, err := net.Dial("unix", "/tmp/"+pluginName+".sock")
	if err != nil {
		return dto.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to connect to plugin: " + pluginName		,
		}
	}
	defer conn.Close()

	requestBytes, _ := json.Marshal(request)
	conn.Write(requestBytes)
	buffer := make([]byte, 32768)
	n, _ := conn.Read(buffer[0:])
	json.Unmarshal(buffer[:n], &response)
	return response
}