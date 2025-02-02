package handler

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"net/url"
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
			return
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

		request := h.createSidraRequest(ctx, route)
		response := h.executePlugins(route.Plugins, request, ctx, startTime, dataplane, gs)
		if response.StatusCode != 0 && response.StatusCode != http.StatusOK {
			return
		}
		for key, value := range response.Headers {
			request.Headers[key] = value
		}
		h.forwardRequest(ctx, request, startTime, dataplane, gs)
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
		for i := len(segments); i > 0; i-- {
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

func (h *Handler) createSidraRequest(ctx *fasthttp.RequestCtx, route dto.SerializeRoute) dto.SidraRequest {
	requestBody := string(ctx.Request.Body())
	request := dto.SidraRequest{
		Upstream: route.UpstreamHost + ":" + route.UpstreamPort,
		Headers:  make(map[string]string),
		Body:     requestBody,
		Url:      string(ctx.Request.URI().RequestURI()),
		Method:   string(ctx.Request.Header.Method()),
		Plugins:  route.Plugins,
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

func (h *Handler) forwardRequest(ctx *fasthttp.RequestCtx, request dto.SidraRequest, startTime time.Time, dataplane, gs string) {
	resp := fasthttp.AcquireResponse()
	h.ForwardToService(ctx, request, resp)
	//@TODO response plugin

	resp.Header.VisitAll(func(key, value []byte) {
		ctx.Response.Header.Set(string(key), string(value))
	})

	ctx.Response.Header.Set("Content-Type", string(resp.Header.Peek("Content-Type")))
	ctx.Response.Header.Set("Server", "Sidra")
	scheme := "http"
	if ctx.IsTLS() {
		scheme = "https"
	}

	// Set Location header to use the scheme from the original request and the path from the upstream response's Location header
	upstreamLocation := string(resp.Header.Peek("Location"))
	if upstreamLocation != "" {
		parsedURL, err := url.Parse(upstreamLocation)
		if err != nil {
			ctx.Response.Header.Set("Location", upstreamLocation)
		}
		location := fmt.Sprintf("%s://%s%s", scheme, ctx.Host(), parsedURL.RequestURI())
		if parsedURL.Fragment != "" {
			location += "#" + parsedURL.Fragment
		}
		ctx.Response.Header.Set("Location", location)
	}
	ctx.Response.SetStatusCode(resp.StatusCode())
	ctx.Response.SetBody(resp.Body())
	h.httpStatusCounter.WithLabelValues(strconv.Itoa(resp.StatusCode()), request.Url, string(ctx.Host()), dataplane, gs).Inc()
	h.requestDuration.WithLabelValues(request.Url, string(ctx.Host()), dataplane, gs).Observe(time.Since(startTime).Seconds())
	fasthttp.ReleaseResponse(resp)
}
