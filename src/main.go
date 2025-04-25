package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

// SearchRequest adalah struktur input API
type SearchRequest struct {
	Target string `json:"target"`
}

// SearchResponse adalah struktur output API
type SearchResponse struct {
	Recipes [][]string `json:"recipes"`
}

var recipes map[string][][]string

// loadRecipes baca recipes.json ke dalam variabel global
func loadRecipes() {
	data, err := os.ReadFile("recipes.json")
	if err != nil {
		log.Fatalf("failed to read recipes.json: %v", err)
	}
	if err := json.Unmarshal(data, &recipes); err != nil {
		log.Fatalf("failed to parse recipes.json: %v", err)
	}
	log.Printf("Loaded %d recipes\n", len(recipes))

	// Debug: Tampilkan semua keys yang tersedia
	log.Println("Available keys in recipes:")
	for key := range recipes {
		log.Printf("- %s\n", key)
	}
}

// searchHandler stub untuk test API
func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received search request")

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight OPTIONS request
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		log.Printf("Method not allowed: %s\n", r.Method)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		log.Printf("Failed to decode request: %v\n", err)
		return
	}

	log.Printf("Searching for target: '%s'\n", req.Target)

	// Untuk test, kembalikan recipes[target] jika ada
	result, ok := recipes[req.Target]
	if !ok {
		log.Printf("Target '%s' not found in recipes\n", req.Target)
		result = [][]string{}
	} else {
		log.Printf("Found %d recipes for target '%s'\n", len(result), req.Target)
	}

	resp := SearchResponse{Recipes: result}
	respData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		log.Printf("Failed to marshal response: %v\n", err)
		return
	}

	log.Printf("Sending response: %s\n", string(respData))
	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}

func main() {
	loadRecipes()

	http.HandleFunc("/api/search", searchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
