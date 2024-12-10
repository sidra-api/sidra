package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	redis "github.com/redis/go-redis/v9"
	"github.com/sidra-gateway/sidra-plugins-hub/lib"
)

type Handler struct {
	redisClient *redis.Client
}

func NewHandler(redisClient *redis.Client) *Handler {
	return &Handler{
		redisClient,
	}
}

func (h *Handler) DefaultHandler() func(w http.ResponseWriter, r *http.Request) {	
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

        key := r.Host + r.URL.Path
		fmt.Println("root key:", key)
        serviceName, err := h.redisClient.HGet(ctx, key, "serviceName").Result()
        if err != nil {
            http.Error(w, "No Route found", http.StatusNotFound)
            return
        }

        servicePort, err := h.redisClient.HGet(ctx, key, "servicePort").Result()
        if err != nil {
            http.Error(w, "No Route found", http.StatusNotFound 	)
            return
        }

		plugins, err := h.redisClient.HGet(ctx, key, "plugins").Result()
        if err != nil {
            plugins = ""
        }		

		// Baca body awal
		bodyBytes, _ := io.ReadAll(r.Body)
		requestBody := string(bodyBytes)
		r.Body = io.NopCloser(strings.NewReader(requestBody)) // Reset body untuk keperluan penerusan berikutnya

		request := lib.SidraRequest{
			Headers: map[string]string{},
			Body:    requestBody,
			Url:     r.URL.String(),
			Method:  r.Method,
		}

		// Salin header
		for key, values := range r.Header {			
			request.Headers[key] = strings.Join(values, " ")
			r.Header.Set(key, strings.Join(values, " "))
		}

		var response lib.SidraResponse

		// Jalankan plugin
		for _, plugin := range strings.Split(plugins, ",") {
			if plugin == "" { continue }
			fmt.Println("Call plugin, name: " + plugin)
			response = h.GoPlugin(plugin, request)		
			for key, values := range response.Headers {		
				r.Header.Set(key, values)	
				request.Headers[key] = values
			}	
			if response.StatusCode != http.StatusOK {
				fmt.Println("Plugin response not OK. Status: ", response.StatusCode)
				w.WriteHeader(response.StatusCode)
				w.Write([]byte(response.Body))
				return
			}
			if w.Header().Get("Cache-Control") != "" && w.Header().Get("Cache-Control") != "no-cache" {
				fmt.Println("ook")
				w.WriteHeader(response.StatusCode)
				w.Write([]byte(response.Body))
			}
		}

		if w.Header().Get("Cache-Control") == "" || w.Header().Get("Cache-Control") == "no-cache" {
			err := h.ForwardToService(w, r, serviceName, servicePort)
			if err != nil {
				http.Error(w, "Failed to forward request to service:"+err.Error(), http.StatusInternalServerError)
			}
		}
		for _, plugin := range strings.Split(plugins, ",") {
			if plugin == "" { continue }			
			response = h.GoPlugin(plugin+".response", request)
		}
	}
}