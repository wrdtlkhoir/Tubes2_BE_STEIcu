package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	// Import package scraper if needed
)

// SearchRequest adalah struktur input API
type SearchRequest struct {
	Target string `json:"target"`
}

// SearchResponse adalah struktur output API
type SearchResponse struct {
	ComponentRecipes [][]string `json:"recipes"`
}

// Global variable for recipe data
var recipeData SimpleOutputData

// Update the loadRecipes function
func loadRecipes(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("failed to read %s: %v", filename, err)
	}
	if err := json.Unmarshal(data, &recipeData); err != nil {
		log.Fatalf("failed to parse %s: %v", filename, err)
	}
	log.Printf("Loaded %d elements and %d recipes from %s\n",
		len(recipeData.Elements), len(recipeData.Recipes), filename)

	// Debug: Tampilkan semua keys yang tersedia
	log.Println("Available elements:")
	for _, elem := range recipeData.Elements {
		log.Printf("- %s\n", elem)
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

	// Look up the component recipes for the requested element
	componentRecipes, ok := recipeData.Recipes[req.Target]
	if !ok {
		log.Printf("Target '%s' not found in recipes\n", req.Target)
		componentRecipes = [][]string{} // Empty slice instead of map
	} else {
		log.Printf("Found recipes for target '%s'\n", req.Target)
	}

	resp := SearchResponse{ComponentRecipes: componentRecipes}
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
	// First scrape the recipes
	var err error
	recipeData, err = ScrapeSimpleRecipes()
	if err != nil {
		log.Fatalf("Error scraping recipes: %v", err)
	}

	// Save them to file
	err = SaveSimpleRecipesToJson(recipeData, "allRecipes.json")
	if err != nil {
		log.Fatalf("Error saving recipes to JSON: %v", err)
	}

	// Alternatively, if you still want to load from file
	// loadRecipes("initial_recipes.json")

	http.HandleFunc("/api/search", searchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
