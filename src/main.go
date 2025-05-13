package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// SearchRequest adalah struktur input API
type SearchRequest struct {
	Target     string `json:"target"`
	Algorithm  string `json:"algorithm"`
	SearchMode string `json:"searchMode"`
	MaxRecipes int    `json:"maxRecipes"`
}

type TreeNode struct {
	Name     string      `json:"name"`
	Children []*TreeNode `json:"children"`
}

type SearchResponse struct {
	Trees         []*TreeNode `json:"tree"`
	NodesVisited  []int       `json:"nodesVisited"`
	ExecutionTime float64     `json:"executionTime"`
}

type MultipleSearchResponse struct {
	Trees         []*TreeNode `json:"trees"`
	NodesVisited  []int       `json:"nodesVisited"`
	ExecutionTime float64     `json:"executionTime"`
}

// Store Recipe Data
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

func convertToTreeNode2(n *Nodebidir) *TreeNode {
	if n == nil {
		return nil
	}

	node := &TreeNode{
		Name:     n.element,
		Children: []*TreeNode{},
	}

	for _, recipe := range n.combinations {
		child1 := convertToTreeNode2(recipe.ingredient1)
		child2 := convertToTreeNode2(recipe.ingredient2)

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

	log.Printf("Searching for target: '%s' using algorithm: %s, mode: %s, maxRecipes: %d\n",
		req.Target, req.Algorithm, req.SearchMode, req.MaxRecipes)

	target := req.Target
	fmt.Printf("Target: %s\n", target)

	startTime := time.Now()

	var resp interface{}
	if req.SearchMode == "single" {
		var node int
		var tree *Tree
		if req.Algorithm == "DFS" {
			tree, node = searchDFSOne(target)
			treeNode := convertToTreeNode(tree.root)
			executionTime := time.Since(startTime).Milliseconds()
			resp = SearchResponse{
				Trees:         []*TreeNode{treeNode},
				ExecutionTime: float64(executionTime),
				NodesVisited:  []int{node},
			}

		} else if req.Algorithm == "BFS" {
			tree, node := searchBFSOne(target)
			treeNode := convertToTreeNode(tree.root)
			executionTime := time.Since(startTime).Milliseconds()
			resp = SearchResponse{
				Trees:         []*TreeNode{treeNode},
				ExecutionTime: float64(executionTime),
				NodesVisited:  []int{node},
			}

		} else {
			tree, node := searchBidirectOne(target)
			treeNode := convertToTreeNode2(tree)
			executionTime := time.Since(startTime).Milliseconds()
			resp = SearchResponse{
				Trees:         []*TreeNode{treeNode},
				ExecutionTime: float64(executionTime),
				NodesVisited:  []int{node},
			}
		}
	} else { // multiple
		var trees []*Tree
		var nodeVisited []int
		if req.Algorithm == "DFS" {
			maxRecipes := req.MaxRecipes
			if maxRecipes <= 1 {
				var node int
				var tree *Tree
				tree, node = searchDFSOne(target)
				treeNode := convertToTreeNode(tree.root)
				executionTime := time.Since(startTime).Milliseconds()
				resp = SearchResponse{
					Trees:         []*TreeNode{treeNode},
					ExecutionTime: float64(executionTime),
					NodesVisited:  []int{node},
				}
			} else {
				trees, nodeVisited = searchDFSMultiple(target, maxRecipes)
				var treeNodes []*TreeNode
				for _, tree := range trees {
					treeNode := convertToTreeNode(tree.root)
					treeNodes = append(treeNodes, treeNode)
				}

				executionTime := time.Since(startTime).Milliseconds()
				resp = MultipleSearchResponse{
					Trees:         treeNodes,
					ExecutionTime: float64(executionTime),
					NodesVisited:  nodeVisited,
				}
			}
		} else if req.Algorithm == "BFS" {
			maxRecipes := req.MaxRecipes
			if maxRecipes <= 1 {
				tree, node := searchBFSOne(target)
				treeNode := convertToTreeNode(tree.root)
				executionTime := time.Since(startTime).Milliseconds()
				resp = SearchResponse{
					Trees:         []*TreeNode{treeNode},
					ExecutionTime: float64(executionTime),
					NodesVisited:  []int{node},
				}
			} else {
				trees, nodeVisited := searchBFSMultiple(target, maxRecipes) //changed to check
				var treeNodes []*TreeNode
				for _, tree := range trees {
					treeNode := convertToTreeNode(tree.root)
					treeNodes = append(treeNodes, treeNode)
				}
				executionTime := time.Since(startTime).Milliseconds()
				resp = MultipleSearchResponse{
					Trees:         treeNodes,
					ExecutionTime: float64(executionTime),
					NodesVisited:  nodeVisited,
				}
			}
		} else {
			maxRecipes := req.MaxRecipes
			if maxRecipes <= 0 {
				maxRecipes = 1 // Default value
			}
			trees, node := searchBidirectionMultiple(target, maxRecipes)
			var treeNodes []*TreeNode
			for _, tree := range trees {
				treeNode := convertToTreeNode2(tree)
				treeNodes = append(treeNodes, treeNode)
			}

			executionTime := time.Since(startTime).Milliseconds()

			// Define a new response structure for multiple trees
			type MultipleSearchResponse struct {
				Trees         []*TreeNode `json:"trees"`
				NodesVisited  int         `json:"nodesVisited"`
				ExecutionTime float64     `json:"executionTime"`
			}

			resp = MultipleSearchResponse{
				Trees:         treeNodes,
				ExecutionTime: float64(executionTime),
				NodesVisited:  node,
			}
		}
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

func main() {
	// First scrape the recipes
	var err error
	recipeData, err = ScrapeRecipes()
	if err != nil {
		log.Fatalf("Error scraping recipes: %v", err)
	}

	// Save them to file
	err = SaveRecipesToJson(recipeData, "recipes.json")
	if err != nil {
		log.Fatalf("Error saving recipes to JSON: %v", err)
	}

	loadRecipes("recipes.json")

	http.HandleFunc("/api/search", searchHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
