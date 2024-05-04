package test

import (
	"fmt"
	"net/http"
)

func Test() {
	// Start four test HTTP servers
	for i := 1; i <= 4; i++ {
		port := 8000 + i
		go startServer(port)
		fmt.Printf("Server %d started on port %d\n", i, port)
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Send the response
		response := fmt.Sprintf("Response from fallback Server %d \n", 8009)
		w.Write([]byte(response))
	})
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8009),
		Handler: handler,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server %d failed to start: %v\n", 8009, err)
	}

	// Keep the main goroutine running
	select {}
}

func startServer(port int) {
	// Create a new ServeMux for this server
	mux := http.NewServeMux()
	// Register handlers
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		// Send the response
		response := fmt.Sprintf("Response from Server %d (Port: %d)\n", port-8000, port)
		w.Write([]byte(response))
	})

	// Start the HTTP server with the custom ServeMux
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Server %d failed to start: %v\n", port-8000, err)
	}
}
