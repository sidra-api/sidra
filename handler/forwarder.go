package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sidra-api/sidra/dto"
	"github.com/valyala/fasthttp"
)

func (h *Handler) ForwardToService(ctx *fasthttp.RequestCtx, request dto.SidraRequest, resp *fasthttp.Response, serviceName, servicePort string)  {
	targetURL := &url.URL{
		Scheme: "http",
		Host:   serviceName + ":" + servicePort,
		Path:   request.Url,
	}

	var client *fasthttp.Client
	readTimeout, _ := time.ParseDuration("500ms")
	writeTimeout, _ := time.ParseDuration("500ms")
	maxIdleConnDuration, _ := time.ParseDuration("1h")
	client = &fasthttp.Client{
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
	for key, value := range request.Headers {
		req.Header.Set(key, value)
	}
	uri := fmt.Sprintf("%s://%s%s", targetURL.Scheme, targetURL.Host, targetURL.Path)
	fmt.Printf("DEBUG Request: %s\n", uri)
	req.SetRequestURI(uri)
	req.Header.SetMethod(request.Method)
	req.SetBodyRaw([]byte(request.Body))	
	err := client.Do(req, resp)
	fasthttp.ReleaseRequest(req)
	if err == nil {
		fmt.Printf("DEBUG Response: %s\n", resp.Body())
	} else {
		fmt.Fprintf(os.Stderr, "ERR Connection error: %v\n", err)
		ctx.Error("Failed to forward request to service", http.StatusInternalServerError)
	}	
}
