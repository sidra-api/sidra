package handler

import (
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	
	"github.com/sidra-api/sidra/dto"	
	"github.com/valyala/fasthttp"	
)

type Handler struct {
	dataSet *dto.DataPlane
}

func NewHandler(dataSet *dto.DataPlane) *Handler {
	return &Handler{
		dataSet,
	}
}

func (h *Handler) DefaultHandler() func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		if (string(ctx.Request.URI().Path()) == "/sidra/healthcheck") {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.Header.Set("Server", "Sidra")
			
			if !processExists("redis") {
				ctx.Response.SetStatusCode(http.StatusInternalServerError)
				return
			}
			ctx.Response.SetStatusCode(http.StatusOK)
			ctx.Response.SetBodyString("{\"status\":\"OK\"}")
			return
		}
		key := string(ctx.Host()) + string(ctx.Request.URI().Path())
		fmt.Println("route key:", key)
		route, exists := h.dataSet.SerializeRoute[key]
		if !exists {
			ctx.Error("Route not found", http.StatusNotFound)
			return
		}
		serviceName := route.UpstreamHost
		servicePort := route.UpstreamPort
		plugins := route.Plugins

		requestBody := string(ctx.Request.Body())

		request := dto.SidraRequest{
			Headers: map[string]string{},
			Body:    requestBody,
			Url:     string(ctx.Request.URI().Path()),
			Method:  string(ctx.Request.Header.Method()),
		}

		// Salin header
		for _, key := range ctx.Request.Header.PeekKeys() {
			k := string(key)
			val := string(ctx.Request.Header.Peek(k))
			request.Headers[k] = val
		}		

		var response dto.SidraResponse

		// Jalankan plugin
		for _, plugin := range strings.Split(plugins, ",") {
			if plugin == "" {
				continue
			}
			fmt.Println("Call plugin, name: " + plugin)
			response = h.GoPlugin(plugin, request)
			for key, values := range response.Headers {
				request.Headers[key] = values
			}
			if response.StatusCode != http.StatusOK {
				fmt.Println("Plugin response not OK. Status: ", response.StatusCode)
				ctx.Error(response.Body, response.StatusCode)
				return
			}
			if response.Headers["Cache-Control"] != "" && response.Headers["Cache-Control"] != "no-cache" {
				ctx.Response.Header.Set("Cache-Control", response.Headers["Cache-Control"])
				ctx.Response.SetBody([]byte(response.Body))
				ctx.Response.SetStatusCode(response.StatusCode)
				return
			}
		}
		resp := fasthttp.AcquireResponse()
		h.ForwardToService(ctx, request, resp, serviceName, servicePort)
				
		for _, plugin := range strings.Split(plugins, ",") {
			if plugin == "" {
				continue
			}
			response = h.GoPlugin(plugin+".response", request)
		}
		for key, values := range response.Headers {
			resp.Header.Set(key, values)
		}
		for _, key := range resp.Header.PeekKeys() {
			k := string(key)
			val := string(resp.Header.Peek(k))
			ctx.Response.Header.Set(k, val)
		}
		ctx.Response.Header.Set("Content-Type", string(resp.Header.Peek("Content-Type")))
		ctx.Response.Header.Set("Server", "Sidra")
		for _, key := range resp.Header.PeekKeys() {
			k := string(key)
			val := string(resp.Header.Peek(k))
			ctx.Response.Header.Set(k, val)
		}
		ctx.Response.SetStatusCode(resp.StatusCode())
		ctx.Response.SetBody(resp.Body())		
		fmt.Println("Response: ", string(resp.Body()))
		fasthttp.ReleaseResponse(resp)	
	}
}

func processExists(processName string) bool {
	cmd := exec.Command("pgrep", "-x", processName)
	output, err := cmd.Output()

	// If `pgrep` finds the process, it returns its PID(s); otherwise, it returns an error.
	if err != nil {
		return false
	}

	// Check if the output is not empty
	return strings.TrimSpace(string(output)) != ""
}