package main

import (
	"log"
	"os"

	"github.com/valyala/fasthttp"
	redis "github.com/redis/go-redis/v9"
	"github.com/sidra-gateway/sidra-plugins-hub/handler"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	h := handler.NewHandler(rdb)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Sidra plugin server is running on port:", port)
	log.Fatal(fasthttp.ListenAndServe(":"+port, h.DefaultHandler()))
}
