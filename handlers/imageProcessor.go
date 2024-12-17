package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go-gemini/processor"
)

// ImageProcessHandler handles image upload and processing
func ImageProcessorHandler(w http.ResponseWriter, r *http.Request) {
	// Validate POST request
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get ID from query parameter
	apiKey := r.URL.Query().Get("api_key")

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	uploadedFile, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer uploadedFile.Close()

	//Call processor
	fmt.Println("Processing image")
	alias, error := processor.StartFromFile(uploadedFile, apiKey)
	if error != nil {
		http.Error(w, fmt.Sprintf("%v", error), http.StatusInternalServerError)
		return
	}

	//Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"alias": alias,
	})
}
