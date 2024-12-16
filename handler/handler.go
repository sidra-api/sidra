package handler

import (	
	"fmt"	
	"net/http"
	"strings"

	redis "github.com/redis/go-redis/v9"
	"github.com/sidra-gateway/sidra-plugins-hub/lib"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	redisClient *redis.Client
}

func NewHandler(redisClient *redis.Client) *Handler {
	return &Handler{
		redisClient,
	}
}

func (h *Handler) DefaultHandler() func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		if (string(ctx.Request.URI().Path()) == "/sidra/healthcheck") {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.Header.Set("Server", "Sidra")
			ctx.Response.SetStatusCode(http.StatusOK)
			ctx.Response.SetBodyString("{\"status\":\"OK\"}")
			return
		}
		key := string(ctx.Host()) + string(ctx.Request.URI().Path())
		fmt.Println("root key:", key)
		serviceName, err := h.redisClient.HGet(ctx, key, "serviceName").Result()
		if err != nil {
			ctx.Error("No Route found", http.StatusNotFound)
			return
		}

		servicePort, err := h.redisClient.HGet(ctx, key, "servicePort").Result()
		if err != nil {
			ctx.Error("No Route found", http.StatusNotFound)
			return
		}

		plugins, err := h.redisClient.HGet(ctx, key, "plugins").Result()
		if err != nil {
			plugins = ""
		}

		requestBody := string(ctx.Request.Body())

		request := lib.SidraRequest{
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

		var response lib.SidraResponse

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
