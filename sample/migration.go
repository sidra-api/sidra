package main

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)
func main() {
	rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    })	
	rdb.HSet(context.Background(), "test.sh:3080/test", "serviceName", "localhost", "servicePort", "8080", "plugins", "bar,foo")
}