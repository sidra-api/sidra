package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/sidra-api/sidra/dto"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	dataSet *dto.DataPlane
}

func NewHandler(dataSet *dto.DataPlane) *Handler {
	return &Handler{
		dataSet: dataSet,
	}
}

func (h *Handler) DefaultHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Request.URI().Path()) == "/sidra/healthcheck" {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.Header.Set("Server", "Sidra")
			ctx.Response.SetStatusCode(http.StatusOK)
			ctx.Response.SetBodyString("{\"status\":\"OK\"}")
			return
		}

		requestPath := string(ctx.Request.URI().Path())
		key := string(ctx.Host()) + requestPath
		fmt.Println("key", key)
		route, exists := h.dataSet.SerializeRoute[key]
		log.Default().Println("-------------BEGIN---------------")

		if requestPath != "/" {
			segments := strings.Split(requestPath, "/")
			for i := 1; i <= len(segments); i++ {
				path := strings.Join(segments[:i], "/")
				if path == "" {
					path = "/"
				}
				fmt.Println("path:", path, string(ctx.Host())+path)
				r := h.dataSet.SerializeRoute[string(ctx.Host())+path]
				fmt.Println("Route:", r)
				if r.PathType == "prefix" {
					route = r
					fmt.Println("Route found: ", route)
					exists = true
					break
				}
			}
		}
		fmt.Println("route", route, exists)
		fmt.Println("All routes", h.dataSet.SerializeRoute)
		if !exists {
			ctx.Error("Route not found", http.StatusNotFound)
			return
		}

		serviceName := route.UpstreamHost
		servicePort := route.UpstreamPort
		plugins := route.Plugins

		requestBody := string(ctx.Request.Body())

		request := dto.SidraRequest{
			Headers: make(map[string]string),
			Body:    requestBody,
			Url:     string(ctx.Request.URI().Path()),
			Method:  string(ctx.Request.Header.Method()),
		}
		// Get client IP address
		clientIP := strings.Split(ctx.RemoteAddr().String(), ":")[0]
		request.Headers["X-Real-Ip"] = clientIP
		// Copy headers
		ctx.Request.Header.VisitAll(func(key, value []byte) {
			request.Headers[string(key)] = string(value)
		})

		var response dto.SidraResponse

		// Execute plugins
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

		resp.Header.VisitAll(func(key, value []byte) {
			ctx.Response.Header.Set(string(key), string(value))
		})

		ctx.Response.Header.Set("Content-Type", string(resp.Header.Peek("Content-Type")))
		ctx.Response.Header.Set("Server", "Sidra")
		ctx.Response.SetStatusCode(resp.StatusCode())
		ctx.Response.SetBody(resp.Body())
		log.Default().Println("--------------END---------------")
		fasthttp.ReleaseResponse(resp)
	}
}
