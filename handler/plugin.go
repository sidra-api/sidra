package handler

import (
	"encoding/json"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sidra-api/sidra/dto"
	"github.com/valyala/fasthttp"
)

func executePlugins(plugins []string, request dto.SidraRequest, ctx *fasthttp.RequestCtx, startTime time.Time, dataplane, gs string) dto.SidraResponse {
	var response dto.SidraResponse

	for _, plugin := range plugins {
		if plugin == "" {
			continue
		}

		response = callPlugin(plugin, request)

		if response.StatusCode != 0 && response.StatusCode != http.StatusOK {
			h.httpStatusCounter.WithLabelValues(strconv.Itoa(response.StatusCode), request.Url, string(ctx.Host()), dataplane, gs).Inc()
			h.requestDuration.WithLabelValues(request.Url, string(ctx.Host()), dataplane, gs).Observe(time.Since(startTime).Seconds())
			ctx.Error(response.Body, response.StatusCode)
			return response
		}

		if response.Headers["Cache-Control"] != "" && response.Headers["Cache-Control"] != "no-cache" {
			ctx.Response.Header.Set("Cache-Control", response.Headers["Cache-Control"])
			h.httpStatusCounter.WithLabelValues(strconv.Itoa(response.StatusCode), request.Url, string(ctx.Host()), dataplane, gs).Inc()
			h.requestDuration.WithLabelValues(request.Url, string(ctx.Host()), dataplane, gs).Observe(time.Since(startTime).Seconds())
			ctx.Response.SetBody([]byte(response.Body))
			ctx.Response.SetStatusCode(response.StatusCode)
			return response
		}
	}

	return response
}

func callPlugin(pluginName string, request dto.SidraRequest) (response dto.SidraResponse) {
	//@TODO: for now body ignored
	request.Body = ""
	conn, err := net.Dial("unix", "/tmp/"+pluginName+".sock")
	if err != nil {
		return dto.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to connect to plugin: " + pluginName,
		}
	}
	defer conn.Close()
	maxMessageSize := 16384
	if size, err := strconv.Atoi(os.Getenv("MAX_MESSAGE_SIZE")); err == nil {
		maxMessageSize = size
	}
	requestBytes, _ := json.Marshal(request)
	conn.Write(requestBytes)
	buffer := make([]byte, maxMessageSize)
	n, _ := conn.Read(buffer[0:])
	json.Unmarshal(buffer[:n], &response)
	return response
}

func callPluginWithBody(pluginName string, request dto.SidraRequest) (response dto.SidraResponse) {
	conn, err := net.Dial("unix", "/tmp/"+pluginName+".sock")
	if err != nil {
		return dto.SidraResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Failed to connect to plugin: " + pluginName,
		}
	}
	defer conn.Close()
	maxMessageSize := 16384
	if size, err := strconv.Atoi(os.Getenv("MAX_MESSAGE_SIZE")); err == nil {
		maxMessageSize = size
	}
	requestBytes, _ := json.Marshal(request)
	conn.Write(requestBytes)
	buffer := make([]byte, maxMessageSize)
	n, _ := conn.Read(buffer[0:])
	json.Unmarshal(buffer[:n], &response)
	return response
}
