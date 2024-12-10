package main

import (	
	"log"
	"net/http"	
	"os"
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
	http.HandleFunc("/", h.DefaultHandler())

	port := os.Getenv("PORT")
	if port == "" {
		port = "3080"
	}

	log.Println("Sidra plugin server is running on port :", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
