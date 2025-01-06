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
	version := "v1.0.0"
	fmt.Println("Sidra plugin server is running  " + version)
	dataSet := dto.NewDataPlane()
	h := handler.NewHandler(dataSet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go func ()  {
		job := scheduler.NewJob(dataSet)
		job.InitialRun()
		job.Run()
	}()
	log.Println("Sidra plugin server is running on port:", port)
	log.Fatal(fasthttp.ListenAndServe(":"+port, h.DefaultHandler()))
}
