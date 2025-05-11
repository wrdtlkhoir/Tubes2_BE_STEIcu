package main

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

// bfsOne performs a standard BFS search (not bidirectional)
// to find a constructible non-cyclic path from the root to base elements
func bfsOne(tree *Treebidir, recipesForItem map[string][][]string) *Nodebidir {
	if tree == nil || tree.root == nil {
		fmt.Println("Tree is empty, cannot perform search.")
		return nil
	}
	if tree.root.isCycleNode {
		fmt.Println("Root is a cycle node, cannot perform search.")
		return nil
	}

	// BFS queue and visited map
	q := list.New()
	visited := make(map[*Nodebidir]*Nodebidir) // Store parent in path: child -> parent

	// Initialize with root
	root := tree.root
	q.PushBack(root)
	visited[root] = nil

	// Track depth for debugging/reporting
	depth := make(map[*Nodebidir]int)
	depth[root] = 0

	fmt.Println("\n--- Starting BFS Search (seeking first constructible non-cyclic path) ---")

	// Perform BFS
	for q.Len() > 0 {
		// Get the front element
		frontElement := q.Front()
		current := frontElement.Value.(*Nodebidir)
		q.Remove(frontElement)

		fmt.Printf("Processing node %s (Depth: %d)\n", current.element, depth[current])

		// Skip cycle nodes
		if current.isCycleNode {
			fmt.Printf("Skipping cycle node %s\n", current.element)
			continue
		}

		// Check if we've reached a base element
		if isBase(current.element) {
			fmt.Printf("Found base element: %s\n", current.element)
			// Now we need to build a path from root to this base element
			pathTree := constructPathTree(current, visited, recipesForItem)
			if pathTree != nil {
				fmt.Printf("Successfully constructed a path tree to base element %s. Returning.\n", current.element)
				return pathTree
			} else {
				fmt.Printf("Path construction failed for base element %s.\n", current.element)
				// Continue the search to find another path
			}
		}

		// Process all combinations (ingredients)
		for _, recipe := range current.combinations {
			children := []*Nodebidir{recipe.ingredient1, recipe.ingredient2}
			for _, child := range children {
				if child == nil || child.isCycleNode {
					continue
				}
				if _, found := visited[child]; !found {
					fmt.Printf("Enqueueing child %s from %s\n", child.element, current.element)
					q.PushBack(child)
					visited[child] = current
					depth[child] = depth[current] + 1
				}
			}
		}
	}

	fmt.Println("--- BFS Search Ended: No constructible non-cyclic path found ---")
	return nil
}

// Constructs a path tree from the target node to the root
func constructPathTree(targetNode *Nodebidir, visited map[*Nodebidir]*Nodebidir, recipeData map[string][][]string) *Nodebidir {
	// First, trace path from target to root
	path := []*Nodebidir{}
	curr := targetNode
	for curr != nil {
		path = append([]*Nodebidir{curr}, path...) // Add to front
		curr = visited[curr]
	}

	fmt.Println("\n--- Path Found ---")
	fmt.Print("Path: ")
	for i, node := range path {
		fmt.Print(node.element)
		if i < len(path)-1 {
			fmt.Print(" -> ")
		}
	}
	fmt.Println()

	// Now build a new tree representing just this path
	return buildShortestPathTree(path, recipeData)
}

// Function to use bfsOne in main
func mainWithBFS(recipeData map[string][][]string) {
	targetElement := "Human" // Example target element
	// Build tree from data
	fullTree := buildTreeBFS(targetElement, recipeData)

	// Print the full tree
	fmt.Println("\nFull Recipe Derivation Tree:")
	printTreeBidir(fullTree)

	// Perform standard BFS search
	num := 3
	pathTree := multipleBfs(fullTree, num)

	if pathTree != nil {
		fmt.Println("\nPath Tree from BFS:")
		fmt.Printf("\nFound %d paths:\n", len(pathTree))
		for i, path := range pathTree {
			fmt.Printf("\nPath %d:\n", i+1)
			printShortestPathTree(path, "", true)
		}

	} else {
		fmt.Println("Could not find a valid path without cycles.")
	}

	fmt.Println("\n" + strings.Repeat("=", 40)) // Separator
}

func multipleBfs(tree *Treebidir, numPaths int) []*Nodebidir {
	if tree == nil || tree.root == nil {
		fmt.Println("Tree is empty, cannot perform search.")
		return nil
	}

	if tree.root.isCycleNode {
		fmt.Println("Root is a cycle node, cannot perform search.")
		return nil
	}

	fmt.Printf("\n--- Starting Multi-Threaded BFS Search for %d Paths ---\n", numPaths)

	// Find all base leaves in the tree
	baseLeaves := findBaseLeaves(tree.root, []*Nodebidir{})
	if len(baseLeaves) == 0 {
		fmt.Println("No base leaves found in the tree, cannot perform search.")
		return nil
	}

	// Filter out cyclic leaf nodes before starting search
	var validLeaves []*Nodebidir
	for _, leaf := range baseLeaves {
		if !leaf.isCycleNode {
			validLeaves = append(validLeaves, leaf)
		} else {
			fmt.Printf("Skipping cyclic leaf node: %s\n", leaf.element)
		}
	}

	if len(validLeaves) == 0 {
		fmt.Println("No valid non-cyclic base leaves found in the tree, cannot perform search.")
		return nil
	}

	fmt.Printf("Found %d valid non-cyclic base leaves as targets: ", len(validLeaves))
	for i, leaf := range validLeaves {
		fmt.Print(leaf.element)
		if i < len(validLeaves)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println()

	resultChan := make(chan PathResult, numPaths*2) // Buffer to avoid deadlocks
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	var foundPaths []*Nodebidir
	var numFoundPaths int

	// Map to track path signatures we've already added
	pathSignatures := make(map[string]bool)

	// Start concurrent searches targeting each valid leaf
	for i, targetLeaf := range validLeaves {
		wg.Add(1)
		go func(leafIndex int, targetLeaf *Nodebidir) {
			defer wg.Done()

			// Check if we should continue searching based on paths found so far
			resultsMutex.Lock()
			shouldContinue := numFoundPaths < numPaths
			resultsMutex.Unlock()

			if !shouldContinue {
				return
			}

			// Skip cyclic leaves (safety check)
			if targetLeaf.isCycleNode {
				fmt.Printf("Skipping cyclic leaf: %s (safety check)\n", targetLeaf.element)
				return
			}

			// Perform a single-threaded BFS from root to this target leaf
			path, score := bfsToTarget(tree.root, targetLeaf)

			// Only proceed if we found a path and it doesn't contain any cyclic nodes
			if path != nil && !isPathCyclic(path) {
				// Generate a unique signature for this path
				pathSignature := generatePathSignature(path)
				desc := fmt.Sprintf("Path to %s (score: %d)", targetLeaf.element, score)

				resultChan <- PathResult{
					path:          path,
					score:         score,
					desc:          desc,
					pathSignature: pathSignature,
				}
			} else if path != nil {
				// Debug output for paths containing cyclic nodes
				fmt.Printf("Found path to %s with cyclic nodes (discarded)\n", targetLeaf.element)
			}
		}(i, targetLeaf)
	}

	// Create a separate goroutine to close the channel after all workers are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Process results as they come in
	for result := range resultChan {
		resultsMutex.Lock()

		// Double-check that the path doesn't contain any cyclic nodes
		if isPathCyclic(result.path) {
			fmt.Printf("Caught cyclic path at collection phase: %s\n", result.desc)
			resultsMutex.Unlock()
			continue
		}

		// Check if we've already found this path by checking its signature
		if numFoundPaths < numPaths && !pathSignatures[result.pathSignature] {
			// This is a new path, add it to our results
			foundPaths = append(foundPaths, result.path)
			numFoundPaths++

			// Mark this path signature as found
			pathSignatures[result.pathSignature] = true

			fmt.Printf("Found path %d/%d: %s\n", numFoundPaths, numPaths, result.desc)
		}
		resultsMutex.Unlock()

		if numFoundPaths >= numPaths {
			break
		}
	}

	fmt.Printf("--- Completed Multi-Threaded BFS Search, found %d paths ---\n", len(foundPaths))
	return foundPaths
}

func bfsToTarget(root *Nodebidir, targetLeaf *Nodebidir) (*Nodebidir, int) {
	if root == nil || targetLeaf == nil {
		return nil, -1
	}

	// Skip if root or target is a cyclic node
	if root.isCycleNode || targetLeaf.isCycleNode {
		return nil, -1
	}

	// Use BFS to find a path from root to the target leaf
	queue := list.New()
	visited := make(map[*Nodebidir]*Nodebidir) // Node -> Parent map for path reconstruction
	queue.PushBack(root)
	visited[root] = nil
	depth := make(map[*Nodebidir]int)
	depth[root] = 0

	foundTarget := false

	// Standard BFS
	for queue.Len() > 0 && !foundTarget {
		current := queue.Front().Value.(*Nodebidir)
		queue.Remove(queue.Front())

		// Skip cyclic nodes
		if current.isCycleNode {
			continue
		}

		// Check if we've reached our target
		if current == targetLeaf {
			foundTarget = true
			break
		}

		// Explore all combinations (children)
		for _, recipe := range current.combinations {
			children := []*Nodebidir{recipe.ingredient1, recipe.ingredient2}
			for _, child := range children {
				// Skip nil or cyclic child nodes
				if child == nil || child.isCycleNode {
					continue
				}

				// Skip already visited nodes
				if _, found := visited[child]; !found {
					queue.PushBack(child)
					visited[child] = current
					depth[child] = depth[current] + 1

					// Early exit if we found our target
					if child == targetLeaf {
						foundTarget = true
						break
					}
				}
			}
			if foundTarget {
				break
			}
		}
	}

	if !foundTarget {
		return nil, -1 // No path found
	}

	// Construct the path tree from root to target
	recipesMapConverted := map[string][][]string(recipeDatas.Recipes)
	resultTree := constructPathTree(targetLeaf, visited, recipesMapConverted)

	// Return path tree and total depth (score)
	return resultTree, depth[targetLeaf]
}

// func main() {

// 	loadRecipes("filtered-recipe.json")
// 	target := "Brick"
// 	numOfRecipe := 2

// 	// ini buat debug result aja
// 	tree := InitTree(target, recipeData.Recipes[target])
// 	printTree(tree)

// 	// Try Single Recipe
// 	result, nodes := searchBFSOne(target)
// 	printTree(result)
// 	fmt.Printf("Number of visited nodes: %d\n", nodes)

// 	// Try multiple Recipe
// 	result2, nodes2 := searchBFSMultiple(target, numOfRecipe)
// 	for _, recipe := range result2 {
// 		printTree(recipe)
// 	}
// 	fmt.Printf("Number of visited nodes: %d\n", nodes2)

// 	// Konversi tree ke JSON
// 	// treeJSON := convertToJSON(tree.root)

// 	// Encode tree ke JSON dan cetak ke stdout
// 	// jsonData, err := json.MarshalIndent(treeJSON, "", "    ")
// 	// if err != nil {
// 	//     log.Fatalf("Failed to encode tree to JSON: %v", err)
// 	// }

// 	// fmt.Println(string(jsonData))
// }
