package handler

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sidra-api/sidra/dto"
	"github.com/valyala/fasthttp"
)

func (h *Handler) forwardRequest(ctx *fasthttp.RequestCtx, request dto.SidraRequest, plugins []string) int {
	resp := fasthttp.AcquireResponse()
	exchange(ctx, request, resp)

	for _, plugin := range plugins {
		if strings.HasPrefix(plugin, "cache") {
			callPluginWithBody(plugin, request)
		}
	}

	resp.Header.VisitAll(func(key, value []byte) {
		ctx.Response.Header.Set(string(key), string(value))
	})

	ctx.Response.Header.Set("Content-Type", string(resp.Header.Peek("Content-Type")))
	ctx.Response.Header.Set("Server", "Sidra")

	scheme := "http"
	if ctx.IsTLS() {
		scheme = "https"
	}

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

	httpStatusCode := resp.StatusCode()
	ctx.Response.SetStatusCode(resp.StatusCode())
	ctx.Response.SetBody(resp.Body())
	fasthttp.ReleaseResponse(resp)

	return httpStatusCode
}

func exchange(ctx *fasthttp.RequestCtx, request dto.SidraRequest, resp *fasthttp.Response) {
	readTimeout, _ := time.ParseDuration("30000ms")
	writeTimeout, _ := time.ParseDuration("30000ms")
	maxIdleConnDuration, _ := time.ParseDuration("1h")
	client := &fasthttp.HostClient{
		Addr:                          request.Upstream,
		ReadTimeout:                   readTimeout,
		WriteTimeout:                  writeTimeout,
		MaxIdleConnDuration:           maxIdleConnDuration,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		// increase DNS cache time to an hour instead of default minute
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}

	req := fasthttp.AcquireRequest()
	if ctx.IsTLS() {
		req.Header.Set("X-Forwarded-Proto", "https")
	} else {
		req.Header.Set("X-Forwarded-Proto", "http")
	}
	req.SetRequestURI(string(ctx.Request.URI().RequestURI()))
	for k, v := range request.Headers {
		if v == "" {
			continue
		}
		req.Header.Set(k, v)
	}
	req.SetBody([]byte(request.Body))
	req.Header.SetMethod(string(request.Method))
	req.Header.SetHost(string(ctx.Host()))
	err := client.Do(req, resp)
	fasthttp.ReleaseRequest(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERR Connection error: %v\n", err)
		resp.SetStatusCode(fasthttp.StatusInternalServerError)
		resp.SetBodyRaw([]byte("upstream error"))
	}
}
