package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
		if h.isHealthCheckRequest(ctx) {
			h.handleHealthCheck(ctx)
			return
		}

		if h.isEndpointsRequest(ctx) {
			h.handleEndpoints(ctx)
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

		httpStatusCode := h.forwardRequest(ctx, request, route.Plugins)
		h.httpStatusCounter.WithLabelValues(strconv.Itoa(httpStatusCode), request.Url, string(ctx.Host()), dataplane, gs).Inc()
		h.requestDuration.WithLabelValues(request.Url, string(ctx.Host()), dataplane, gs).Observe(time.Since(startTime).Seconds())
	}
}

func (h *Handler) isHealthCheckRequest(ctx *fasthttp.RequestCtx) bool {
	return string(ctx.Request.URI().Path()) == "/sidra/healthcheck"
}

func (h *Handler) isEndpointsRequest(ctx *fasthttp.RequestCtx) bool {
	return string(ctx.Request.URI().Path()) == "/sidra/endpoints" && h.isLocalRequest(ctx)
}

func (h *Handler) isLocalRequest(ctx *fasthttp.RequestCtx) bool {
	remoteAddr := ctx.RemoteAddr().String()
	return strings.HasPrefix(remoteAddr, "127.0.0.1") || strings.HasPrefix(remoteAddr, "[::1]")
}

func (h *Handler) handleHealthCheck(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.Header.Set("Server", "Sidra")
	ctx.Response.SetStatusCode(http.StatusOK)
	ctx.Response.SetBodyString("{\"status\":\"OK\"}")
}

func (h *Handler) handleEndpoints(ctx *fasthttp.RequestCtx) {
	jsonResponse, _ := json.Marshal(h.dataSet.SerializeRoute)
	log.Default().Println("endpoints", string(jsonResponse))
	ctx.Response.SetStatusCode(http.StatusOK)
	ctx.Response.SetBody(jsonResponse)
}
