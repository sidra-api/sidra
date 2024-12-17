package main

import (
	"log"
	"os"

	"github.com/sidra-api/sidra/dto"
	"github.com/sidra-api/sidra/handler"
	"github.com/sidra-api/sidra/scheduler"
	"github.com/valyala/fasthttp"
)

func main() {
	dataSet := dto.NewDataPlane()
	h := handler.NewHandler(dataSet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	go func ()  {
		scheduler.NewJob(dataSet).Run()
	}()
	log.Println("Sidra plugin server is running on port:", port)
	log.Fatal(fasthttp.ListenAndServe(":"+port, h.DefaultHandler()))
}
