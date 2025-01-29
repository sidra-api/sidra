package handler

import (
	"fmt"
	"os"
	"time"

	"github.com/sidra-api/sidra/dto"
	"github.com/valyala/fasthttp"
)

func (h *Handler) ForwardToService(ctx *fasthttp.RequestCtx, request dto.SidraRequest, resp *fasthttp.Response) {
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
		fmt.Printf("DEBUG Header: %s: %s\n", k, v)
		req.Header.Add(k, v)
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
