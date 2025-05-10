package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// Data structures for recipes
// SearchRequest is the API input structure
type SearchRequest struct {
	Target string `json:"target"`
}

// SearchResponse is the API output structure
type SearchResponse struct {
	ComponentRecipes [][]string `json:"recipes"`
}

// SearchResponseRec is the recursive API output structure
type SearchResponseRec struct {
	ComponentRecipesRec map[string][][]string `json:"recipes"`
}

// FullSearchResponse contains all response data including shortest path
type FullSearchResponse struct {
	ComponentRecipes    [][]string            `json:"recipes"`
	ComponentRecipesRec map[string][][]string `json:"recursive_recipes"`
	ShortestPath        [][]string            `json:"shortest_path"`
}

// Global variable for recipe data
var recipeData SimpleOutputData
var recipeDatas OutputData

// searchHandler handles the API requests
func searchHandler(w http.ResponseWriter, r *http.Request) {
	// Setup headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid input"}`, http.StatusBadRequest)
		return
	}

	// Extract target from the request
	target := req.Target
	log.Printf("Received search request for target: %s", target)

	// Recipe lookups from simple recipes
	componentRecipes, ok1 := recipeData.Recipes[target]
	if !ok1 {
		componentRecipes = [][]string{}
	}

	// Extract the recursive recipes for the target
	componentRecipesRec := make(map[string][][]string)
	if targetRecipes, ok := recipeDatas.Recipes[target]; ok {
		// Extract all recipes for the target
		for key, recipes := range targetRecipes {
			componentRecipesRec[key] = recipes
		}
	}

	// Create a simplified recipe map for the BiMap
	// We need to extract map[string][][]string from the initial data
	subMap := make(map[string][][]string)

	// First add the target recipes directly
	if targetMap, ok := recipeDatas.Recipes[target]; ok {
		for _, recipes := range targetMap {
			// Add all recipes for this target
			if subMap[target] == nil {
				subMap[target] = [][]string{}
			}
			subMap[target] = append(subMap[target], recipes...)
		}
	}

	// Find shortest path using BiMap
	// bi := NewBiMap(subMap)
	// path := bi.FindShortestRecipe(target)

	// Build full response
	resp := FullSearchResponse{
		ComponentRecipes:    componentRecipes,
		ComponentRecipesRec: componentRecipesRec,
		// ShortestPath:        path,
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(respData)
}

// func main() {
// 	// First scrape the recipes
// 	var err error
// 	var err2 error

// 	log.Println("Starting to scrape simple recipes...")
// 	recipeData, err = ScrapeSimpleRecipes()
// 	if err != nil {
// 		log.Fatalf("Error scraping simple recipes: %v", err)
// 	}

// 	log.Println("Starting to scrape recursive recipes...")
// 	recipeDatas, err2 = ScrapeInitialRecipes()
// 	if err2 != nil {
// 		log.Fatalf("Error scraping recursive recipes: %v", err2)
// 	}

// 	// Save them to file
// 	log.Println("Saving simple recipes to JSON...")
// 	err = SaveSimpleRecipesToJson(recipeData, "allRecipes.json")
// 	if err != nil {
// 		log.Fatalf("Error saving simple recipes to JSON: %v", err)
// 	}

// 	log.Println("Saving recursive recipes to JSON...")
// 	err2 = SaveRecipesToJson(recipeDatas, "initial_recipes.json")
// 	if err2 != nil {
// 		log.Fatalf("Error saving recursive recipes to JSON: %v", err2)
// 	}

// 	// Setup HTTP server
// 	http.HandleFunc("/api/search", searchHandler)

// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080"
// 	}

// 	log.Printf("Server running on port %s\n", port)
// 	log.Printf("Ready to serve recipe requests!")
// 	log.Fatal(http.ListenAndServe(":"+port, nil))
// }
