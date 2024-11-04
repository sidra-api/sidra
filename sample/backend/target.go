package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func echoHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	fmt.Fprintf(w, "Request received at terget server : \nPath: %s\nMethod: %s\nBody: %s", 
r.URL.Path, r.Method, string(body))
}

func main() {
	http.HandleFunc("/", echoHandler)
	log.Println("Target server running on port 7070.")
	log.Fatal(http.ListenAndServe(":7070", nil))
}