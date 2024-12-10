package main

import (
	"fmt"
	"time"

	redis "github.com/redis/go-redis/v9"
	"github.com/sidra-gateway/sidra-plugins-hub/handler"
	"github.com/sidra-gateway/sidra-plugins-hub/lib"
)


func main() {
	rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })
	h := handler.NewHandler(rdb)	
	
	go func() {
		fmt.Println("Starting date time:", time.Now())
		r2 := h.GoPlugin("foo", lib.SidraRequest{
			Headers: map[string]string{"test":"1"},
		})
		fmt.Println(r2.Headers["test"])
		if r2.Headers["test"] == "1" {
			fmt.Println("Success")
		} else {
			fmt.Println("Failed")
		}		
	}()
	fmt.Println("Starting date time:", time.Now())
	r1 := h.GoPlugin("foo", lib.SidraRequest{
		Headers: map[string]string{"test":"2"},
	})
	fmt.Println(r1.Headers["test"])
	if r1.Headers["test"] == "2" {
		fmt.Println("Success")
	} else {
		fmt.Println("Failed")
	}	

	
	
}