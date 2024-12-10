package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (h *Handler) ForwardToService(w http.ResponseWriter, r *http.Request, serviceName, servicePort string) error {
	targetURL := &url.URL{
		Scheme: "http",
		Host:   serviceName + ":" + servicePort,
		Path:   r.URL.Path,
	}

	// Re-read body to forward it to the service
	bodyBytes, _ := io.ReadAll(r.Body)

	r.Body = io.NopCloser(strings.NewReader(string(bodyBytes))) // Reset body for future use in plugins

	proxyReq, _ := http.NewRequest(r.Method, targetURL.String(), io.NopCloser(strings.NewReader(string(bodyBytes))))

	// Copy headers from original request to new request
	for key, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Send the request to the target service
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write response headers and status code from target service to response writer
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}	
	io.Copy(w, resp.Body)
	return nil
}