package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sidra-api/sidra/dto"
	"github.com/sidra-api/sidra/handler"
	"github.com/sidra-api/sidra/scheduler"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	httpStatusCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_status_total",
			Help: "Count of HTTP response statuses",
		},
		[]string{"code", "path", "host", "dataplane", "gs"},
	)
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "host", "dataplane", "gs"},
	)
)

func init() {
	prometheus.MustRegister(httpStatusCounter)
	prometheus.MustRegister(requestDuration)
}

func main() {
	version := "v1.0.1"
	fmt.Println("Sidra plugin server is running  " + version)
	dataSet := dto.NewDataPlane()
	h := handler.NewHandler(dataSet, httpStatusCounter, requestDuration)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	portSll := os.Getenv("SSL_PORT")
	if portSll == "" {
		portSll = "8443"
	}
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		job := scheduler.NewJob(dataSet)
		job.InitialRun()
		job.Run()
	}()
	go func() {
		certFile := os.Getenv("SSL_CERT_FILE")
		if certFile == "" {
			certFile = "/tmp/server.crt"
		}
		keyFile := os.Getenv("SSL_KEY_FILE")
		if keyFile == "" {
			keyFile = "/tmp/server.key"
		}
		log.Fatal(fasthttp.ListenAndServeTLS("0.0.0.0:"+portSll, certFile, keyFile, h.DefaultHandler()))
	}()
	go func() {
		log.Fatal(fasthttp.ListenAndServe("0.0.0.0:"+port, h.DefaultHandler()))
	}()
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe("0.0.0.0:9100", nil))
	}()
	log.Println("Sidra plugin server is running on port:", port)
	<-shutdownChan
	fmt.Println("\nShutdown signal received")
	cleanPluginSocket(dataSet.Plugins)
}

func cleanPluginSocket(plugins map[string]dto.Plugin) {
	for _, plugin := range plugins {
		file := filepath.Join("/tmp", plugin.Name+".sock")
		err := os.Remove(file)
		if err != nil {
			log.Printf("Failed to remove %s: %v\n", file, err)
		} else {
			fmt.Printf("Removed: %s\n", file)
		}
	}
}
