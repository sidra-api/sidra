package main

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

func forwardRequest(w http.ResponseWriter, r *http.Request) {

	serviceName := r.Header.Get("ServiceName")
	servicePort := r.Header.Get("ServicePort")
	plugins := r.Header.Get("Plugins")

	log.Println("Plugins :", plugins)

	if serviceName == "" || servicePort == "" {
		http.Error(w, "ServiceName or ServicePort are not available.", http.StatusBadRequest)
		return
	}

	targetURL := &url.URL{
		Scheme: "http",
		Host:   serviceName + ":" + servicePort,
		Path:   r.URL.Path,
	}

	req, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
	if err != nil {
		log.Println("Failed to make request :", err)
		http.Error(w, "Failed to make request.", http.StatusInternalServerError)
		return
	}

	req.Header = r.Header

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to forward request :", err)
		http.Error(w, "Failed to forward request.", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func main() {
	http.HandleFunc("/", forwardRequest)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Sidra plugin server running on port :", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
