package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func requestLoggerHandler(w http.ResponseWriter, r *http.Request) {
	// Log request details
	fmt.Printf("[%s] %s %s from %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path, r.RemoteAddr)

	// Read and log the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading body: %v\n", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)

		return
	}

	// Log the body
	if len(body) > 0 {
		fmt.Printf("Body: %s\n", string(body))
	} else {
		fmt.Println("Body: <empty>")
	}

	// Log headers
	// fmt.Println("Headers:")
	// for name, values := range r.Header {
	// 	for _, value := range values {
	// 		fmt.Printf("  %s: %s\n", name, value)
	// 	}
	// }

	fmt.Println("-----------------------------------")

	// Send a simple response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Request logged successfully"))
}

func main() {
	port := "8080"

	// Use custom port if provided as argument
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// Create handler for all paths
	http.HandleFunc("/", requestLoggerHandler)

	// Start the server
	addr := ":" + port
	fmt.Printf("Starting request logger server on %s\n", addr)
	fmt.Println("Press Ctrl+C to stop")

	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
