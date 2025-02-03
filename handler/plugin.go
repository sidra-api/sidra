package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/sidra-api/sidra/dto"
	"github.com/valyala/fasthttp"
)

func (h *Handler) executePlugins(plugins []string, request dto.SidraRequest, ctx *fasthttp.RequestCtx, startTime time.Time, dataplane, gs string) dto.SidraResponse {
	var response dto.SidraResponse

	for _, plugin := range plugins {
		if plugin == "" {
			continue
		}
		dataPlugin := h.dataSet.Plugins[plugin]
		response = callPlugin(plugin, request, dataPlugin)

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

func callPlugin(pluginName string, request dto.SidraRequest, dataPlugin dto.Plugin) (response dto.SidraResponse) {
	//@TODO: for now body ignored
	request.Body = ""
	conn, err := net.Dial("unix", "/tmp/"+pluginName+".sock")
	if err != nil {
		restartPlugin(dataPlugin)
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

func callPluginWithBody(pluginName string, request dto.SidraRequest, dataPlugin dto.Plugin) (response dto.SidraResponse) {
	conn, err := net.Dial("unix", "/tmp/"+pluginName+".sock")
	if err != nil {
		restartPlugin(dataPlugin)
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

func restartPlugin(plugin dto.Plugin) {
	if plugin.Enabled == 0 {
		return
	}
	basePath := "/usr/local/bin/"
	if _, err := os.Stat("/usr/local/bin/" + plugin.TypePlugin); err != nil {
		if _, err := os.Stat("./plugins/plugin-" + plugin.TypePlugin + "/" + plugin.TypePlugin); err != nil {
			fmt.Println("Plugin does not exist")
			return
		} else {
			basePath = "./plugins/plugin-" + plugin.TypePlugin + "/"
		}
	}
	os.Remove("/tmp/" + plugin.Name + ".sock")
	env := make(map[string]string)
	err := json.Unmarshal([]byte(plugin.Config), &env)
	if err != nil {
		fmt.Println("- Failed to parse plugin config", plugin.TypePlugin, plugin.Config, err)
		return
	}
	env["PLUGIN_NAME"] = plugin.Name
	cmd := exec.Command(basePath + plugin.TypePlugin)
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	err = cmd.Start()
	if err != nil {
		fmt.Println("- Failed to start plugin", plugin.TypePlugin, err)
		return
	}
}
