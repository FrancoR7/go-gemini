package main

import (
	"fmt"
	"log"
	"net/http"

	"go-gemini/handlers"
)

func main() {
	// Register handlers
	http.HandleFunc("/imageProcessor", handlers.ImageProcessorHandler)

	// Start server
	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
