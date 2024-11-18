package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println("Error reading body:", err)
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		log.Printf("Request received at target server: \nPath: %s\nMethod: %s\nBody: %s\n", r.URL.Path, r.Method, string(body))

		// Kirim respons
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, "Request received at target server: \nPath: %s\nMethod: %s\nBody: %s\n", r.URL.Path, r.Method, string(body))
	})

	log.Println("Backend server is running on port 7070")
	log.Fatal(http.ListenAndServe(":7070", nil))
}