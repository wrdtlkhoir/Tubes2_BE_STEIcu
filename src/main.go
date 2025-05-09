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
	Algorithm   string `json:"algorithm"` // tambahkan mode
	SearchMode string `json:"searchMode"`
}

type TreeNode struct {
    Name     string     `json:"name"`
    Children []*TreeNode `json:"children"`
}


// SearchResponse adalah struktur output API
// SearchResponse adalah struktur output API
type SearchResponse struct {
	Path [][]string `json:"path"`
	ComponentRecipes map[string][][]string `json:"recipes"`
}

// Change the global variable definition
var recipeData OutputData

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
func convertNodeToPaths(node *Node) [][]string {
    var paths [][]string
    var dfs func(n *Node, path []string)
    dfs = func(n *Node, path []string) {
        if n == nil {
            return
        }
        path = append(path, n.element)
        if isLeaf(n) {
            paths = append(paths, append([]string{}, path...))
            return
        }
        for _, recipe := range n.combinations {
            dfs(recipe.ingredient1, path)
            dfs(recipe.ingredient2, path)
        }
    }
    dfs(node, []string{})
    return paths
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

	// Gunakan dummy data dan DFS untuk pencarian path
	var paths [][]string
    tree := InitTree(req.Target, dummy)
    if req.SearchMode == "shortest" {
        // Ambil semua path pada shortest branch
        shortestPathNode := searchDFSOne(tree)
        paths = convertNodeToPaths(shortestPathNode)
    } //else {
        // Multiple path
        //paths, _ = searchDFSMultiple(10, tree)
    //}

	// Dummy komponen resep (bisa diisi sesuai kebutuhan)
    componentRecipes := make(map[string][][]string)
    if recipes, ok := dummy[req.Target]; ok {
        componentRecipes[req.Target] = recipes
    }

    resp := SearchResponse{
        Path: paths,
        ComponentRecipes: componentRecipes,
    }
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
// Generalisasi fungsi build tree dari dummy (dfs.go)
func buildGeneralTree(target string) *TreeNode {
    visited := make(map[string]bool)
    return buildGeneralTreeHelper(target, visited)
}

func buildGeneralTreeHelper(element string, visited map[string]bool) *TreeNode {
    // Hindari infinite loop jika ada siklus
    if visited[element] {
        return &TreeNode{Name: element, Children: []*TreeNode{}}
    }
    visited[element] = true

    recipes, ok := dummy[element]
    if !ok || len(recipes) == 0 {
        return &TreeNode{Name: element, Children: []*TreeNode{}}
    }

    // Ambil kombinasi pertama (shortest path) saja
    children := []*TreeNode{}
    for _, ingredient := range recipes[0] {
        child := buildGeneralTreeHelper(ingredient, visited)
        children = append(children, child)
    }

    return &TreeNode{
        Name:     element,
        Children: children,
    }
}
// // Fungsi dummy untuk tree "Brick" sesuai gambar
// func buildBrickTree(target string) *TreeNode {
//     if target != "Brick" {
//         return &TreeNode{Name: target, Children: []*TreeNode{}}
//     }
//     return &TreeNode{
//         Name: "Brick",
//         Children: []*TreeNode{
//             {
//                 Name: "Mud",
//                 Children: []*TreeNode{
//                     {Name: "Water", Children: []*TreeNode{}},
//                     {Name: "Earth", Children: []*TreeNode{}},
//                 },
//             },
//             {Name: "Fire", Children: []*TreeNode{}},
//         },
//     }
// }
// Handler endpoint /api/tree
func treeHandler(w http.ResponseWriter, r *http.Request) {
    log.Println("Received tree request")

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
	log.Printf("Building tree for target: '%s'\n", req.Target)

    // Untuk demo, jika target "Brick", kembalikan tree sesuai gambar
    tree := buildGeneralTree(req.Target)

    respData, err := json.Marshal(tree)
    if err != nil {
        http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
        log.Printf("Failed to marshal tree: %v\n", err)
        return
    }

    log.Printf("Sending tree response: %s\n", string(respData))
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
    http.HandleFunc("/api/tree", treeHandler) // Tambahkan endpoint baru

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server running on port %s\n", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}

