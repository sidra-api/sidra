package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sidra-api/sidra/dto"
	"github.com/sidra-api/sidra/handler"
	"github.com/sidra-api/sidra/scheduler"
	"github.com/valyala/fasthttp"
)

func main() {
	version := "v1.0.1"
	fmt.Println("Sidra plugin server is running  " + version)
	dataSet := dto.NewDataPlane()
	h := handler.NewHandler(dataSet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	portSll := os.Getenv("SSL_PORT")
	if portSll == "" {
		portSll = "8443"
	}
	go func() {
		job := scheduler.NewJob(dataSet)
		job.InitialRun()
		job.Run()
	}()
	log.Println("Sidra plugin server is running on port:", port)
	go func() {
		if os.Getenv("SSL_ON") == "true" {
			certFile := os.Getenv("SSL_CERT_FILE")
			if certFile == "" {
				certFile = "/etc/ssl/certs/server.crt"
			}
			keyFile := os.Getenv("SSL_KEY_FILE")
			if keyFile == "" {
				keyFile = "/etc/ssl/private/server.key"
			}
			log.Fatal(fasthttp.ListenAndServeTLS(":"+portSll, certFile, keyFile, h.DefaultHandler()))
		}
	}()
	log.Fatal(fasthttp.ListenAndServe(":"+port, h.DefaultHandler()))

}
