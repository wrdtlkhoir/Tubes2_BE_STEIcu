package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
    "time"
)

// SearchRequest adalah struktur input API
type SearchRequest struct {
	Target      string `json:"target"`
	Algorithm   string `json:"algorithm"`
	SearchMode  string `json:"searchMode"`
    MaxRecipes  int    `json:"maxRecipes"`
}

type TreeNode struct {
    Name     string      `json:"name"`
    Children []*TreeNode `json:"children"`
}

type SearchResponse struct {
	Tree           *TreeNode `json:"tree"`
    NodesVisited     int        `json:"nodesVisited"`
    ExecutionTime    float64      `json:"executionTime"`
}

// Change the global variable definition
var recipeData OutputData

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
	// log.Println("Available elements:")
	// for _, elem := range recipeData.Elements {
	// 	log.Printf("- %s\n", elem)
	// }
}
func convertToTreeNode(n *Node) *TreeNode {
	if n == nil {
		return nil
	}

	node := &TreeNode{
		Name:     n.element,
		Children: []*TreeNode{},
	}

	for _, recipe := range n.combinations {
		child1 := convertToTreeNode(recipe.ingredient1)
		child2 := convertToTreeNode(recipe.ingredient2)

		if child1 != nil {
			node.Children = append(node.Children, child1)
		}
		if child2 != nil {
			node.Children = append(node.Children, child2)
		}
	}

	return node
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received search request")

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

    startTime := time.Now()

    var node int
    var tree *Tree
    // var trees []*Tree
    // var nodes []int

    //temp nanti diganti yg sesuai input user
    // var numOfRecipe int

    if req.SearchMode == "single" {
        if (req.Algorithm == "DFS") {
            tree, node = searchDFSOne(req.Target)
        } else {
            tree, node = searchBFSOne(req.Target)
        }
    } else { // multiple
        // if req.Algorithm == "DFS" {
        //     trees, nodes = searchDFSMultiple(req.Target, numOfRecipe)
        // } else {
        //     trees, nodes = searchBFSMultiple(req.Target, numOfRecipe)
        // }
        // // nambahin convert masing2 tree ke path [][]string
    }
    treeNode := convertToTreeNode(tree.root)
    executionTime := time.Since(startTime).Milliseconds()

    resp := SearchResponse{
        Tree: treeNode,
        ExecutionTime:    float64(executionTime),
        NodesVisited:     node,
    }
    respData, err := json.Marshal(resp)
    if err != nil {
        http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
        log.Printf("Failed to marshal response: %v\n", err)
        return
    }
    printTree(tree) // Debug: Print the tree structure to the console

    log.Printf("Sending response: %s\n", string(respData))
    w.WriteHeader(http.StatusOK)
    w.Write(respData)
}

func main() {
    // First scrape the recipes
    var err error
    recipeData, err = ScrapeInitialRecipes()
    if err != nil {
        log.Fatalf("Error scraping recipes: %v", err)
    }

    // Save them to file
    err = SaveRecipesToJson(recipeData, "initial_recipes.json")
    if err != nil {
        log.Fatalf("Error saving recipes to JSON: %v", err)
    }

    http.HandleFunc("/api/search", searchHandler)
    // http.HandleFunc("/api/tree", treeHandler) // Tambahkan endpoint baru

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server running on port %s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

