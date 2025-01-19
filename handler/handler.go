package handler

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sidra-api/sidra/dto"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	dataSet           *dto.DataPlane
	httpStatusCounter *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
}

func NewHandler(dataSet *dto.DataPlane, httpStatusCounter *prometheus.CounterVec, requestDuration *prometheus.HistogramVec) *Handler {
	return &Handler{
		dataSet:           dataSet,
		httpStatusCounter: httpStatusCounter,
		requestDuration:   requestDuration,
	}
}

func (h *Handler) DefaultHandler() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Request.URI().Path()) == "/sidra/healthcheck" {
			h.handleHealthCheck(ctx)
			return
		}
		if string(ctx.Request.URI().Path()) == "/endpoints" {
			jsonResponse, _ := json.Marshal(h.dataSet.SerializeRoute)
			log.Default().Println("endpoints", string(jsonResponse))
			ctx.Response.SetStatusCode(200)
			ctx.Response.SetBody(jsonResponse)
		}
		startTime := time.Now()
		route, exists := h.findRoute(ctx)
		if !exists {
			ctx.Error("Route not found", http.StatusNotFound)
			return
		}

		dataplane := os.Getenv("dataplaneid")
		if dataplane == "" {
			dataplane = "standalone"
		}
		gs := route.GatewayID
		if gs == "" {
			gs = "standalone"
		}

		request := h.createSidraRequest(ctx)
		response := h.executePlugins(route.Plugins, request, ctx, startTime, dataplane, gs)
		if response.StatusCode != 0 && response.StatusCode != http.StatusOK {
			return
		}

		h.forwardRequest(ctx, request, route, response, startTime, dataplane, gs)
	}
}

func (h *Handler) handleHealthCheck(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.Header.Set("Server", "Sidra")
	ctx.Response.SetStatusCode(http.StatusOK)
	ctx.Response.SetBodyString("{\"status\":\"OK\"}")
}

func (h *Handler) findRoute(ctx *fasthttp.RequestCtx) (dto.SerializeRoute, bool) {
	requestPath := string(ctx.Request.URI().Path())
	key := string(ctx.Host()) + requestPath
	route, exists := h.dataSet.SerializeRoute[key]
	if exists {
		return route, exists
	}
	if requestPath != "/" {
		segments := strings.Split(requestPath, "/")
		for i := 1; i <= len(segments); i++ {
			path := strings.Join(segments[:i], "/")
			if path == "" {
				path = "/"
			}
			r := h.dataSet.SerializeRoute[string(ctx.Host())+path]
			if r.PathType == "prefix" {
				route = r
				exists = true
				break
			}
		}
	}
	return route, exists
}

func (h *Handler) createSidraRequest(ctx *fasthttp.RequestCtx) dto.SidraRequest {
	requestBody := string(ctx.Request.Body())
	request := dto.SidraRequest{
		Headers: make(map[string]string),
		Body:    requestBody,
		Url:     string(ctx.Request.URI().Path()),
		Method:  string(ctx.Request.Header.Method()),
	}
	clientIP := strings.Split(ctx.RemoteAddr().String(), ":")[0]
	request.Headers["X-Real-IP"] = clientIP
	log.Default().Println("ClientIP", clientIP)
	ctx.Request.Header.VisitAll(func(key, value []byte) {
		request.Headers[string(key)] = string(value)
	})
	return request
}

func (h *Handler) executePlugins(plugins []string, request dto.SidraRequest, ctx *fasthttp.RequestCtx, startTime time.Time, dataplane, gs string) dto.SidraResponse {
	var response dto.SidraResponse
	for _, plugin := range plugins {
		if plugin == "" {
			continue
		}
		response = h.GoPlugin(plugin, request)
		for key, values := range response.Headers {
			request.Headers[key] = values
		}
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

func (h *Handler) forwardRequest(ctx *fasthttp.RequestCtx, request dto.SidraRequest, route dto.SerializeRoute, response dto.SidraResponse, startTime time.Time, dataplane, gs string) {
	resp := fasthttp.AcquireResponse()
	h.ForwardToService(ctx, request, resp, route.UpstreamHost, route.UpstreamPort)
	for _, plugin := range route.Plugins {
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
	h.httpStatusCounter.WithLabelValues(strconv.Itoa(response.StatusCode), request.Url, string(ctx.Host()), dataplane, gs).Inc()
	h.requestDuration.WithLabelValues(request.Url, string(ctx.Host()), dataplane, gs).Observe(time.Since(startTime).Seconds())
	fasthttp.ReleaseResponse(resp)
}
