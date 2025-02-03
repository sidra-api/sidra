package handler

import (
	"github.com/sidra-api/sidra/dto"
	"github.com/valyala/fasthttp"
	"log"
	"strings"
)

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
