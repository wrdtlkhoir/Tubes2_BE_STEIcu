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
	targetElement := "Ozone" // Example target element
	// Build tree from data
	fullTree := buildTreeBFS(targetElement, recipeData)

	// Print the full tree
	fmt.Println("\nFull Recipe Derivation Tree:")
	printTreeBidir(fullTree)

	// Perform standard BFS search
	pathTree := bfsOne(fullTree, recipeData)

	if pathTree != nil {
		fmt.Println("\nPath Tree from BFS:")
		printShortestPathTree(pathTree, "", true)
	} else {
		fmt.Println("Could not find a valid path without cycles.")
	}

	fmt.Println("\n" + strings.Repeat("=", 40)) // Separator
}

/*** MULTIPLE RECIPE BFS ***/
func searchBFSMultiple(target string, maxPathsToReturn int) ([]*Tree, []int) {
	fmt.Println("start bfs multiple")
	targetSpecificRecipes, ok := recipeData.Recipes[target]
	if !ok {
		return []*Tree{}, []int{}
	}

	rootNodes := bfsAll(target, maxPathsToReturn, targetSpecificRecipes)

	var trees []*Tree
	var pathElementCounts []int

	for _, rootNode := range rootNodes {
		trees = append(trees, &Tree{root: rootNode})
		pathElementCounts = append(pathElementCounts, getPathElementCount(rootNode))
	}
	return trees, pathElementCounts
}

func bfsAll(targetElement string, maxPathsToReturn int, currentRecipeMap map[string][][]string) []*Node {
	var allFinalTargetTrees []*Node
	var mu sync.Mutex
	var wg sync.WaitGroup

	type explorationNode struct {
		element string
		node    *Node
		depth   int
	}

	targetCombs, exists := currentRecipeMap[targetElement]
	if !exists || len(targetCombs) == 0 {
		return []*Node{{element: targetElement}}
	}

	sem := make(chan struct{}, 8) // limit concurrency to 8 goroutines
	resultChan := make(chan *Node, len(targetCombs))

	for _, pair := range targetCombs {
		if len(pair) != 2 {
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // acquire semaphore
		go func(pair []string) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore

			rootNode := &Node{
				element: targetElement,
				combinations: []Recipe{{
					ingredient1: nil,
					ingredient2: nil,
				}},
			}

			queue := []explorationNode{
				{element: pair[0], node: nil, depth: 1},
				{element: pair[1], node: nil, depth: 1},
			}

			visited := make(map[string]bool)
			visited[targetElement] = true
			elementNodes := make(map[string]*Node)

			leftNode := &Node{element: pair[0]}
			rightNode := &Node{element: pair[1]}
			rootNode.combinations[0].ingredient1 = leftNode
			rootNode.combinations[0].ingredient2 = rightNode

			queue[0].node = leftNode
			queue[1].node = rightNode

			elementNodes[pair[0]] = leftNode
			elementNodes[pair[1]] = rightNode

			idx := 0
			for idx < len(queue) {
				current := queue[idx]
				idx++

				if idx > 2 && visited[current.element] {
					continue
				}
				visited[current.element] = true

				combs, exists := currentRecipeMap[current.element]
				if !exists || len(combs) == 0 || len(combs[0]) != 2 {
					continue
				}

				ingredient1, ingredient2 := combs[0][0], combs[0][1]

				leftIngNode, exists := elementNodes[ingredient1]
				if !exists {
					leftIngNode = &Node{element: ingredient1}
					elementNodes[ingredient1] = leftIngNode
					queue = append(queue, explorationNode{element: ingredient1, node: leftIngNode, depth: current.depth + 1})
				}

				rightIngNode, exists := elementNodes[ingredient2]
				if !exists {
					rightIngNode = &Node{element: ingredient2}
					elementNodes[ingredient2] = rightIngNode
					queue = append(queue, explorationNode{element: ingredient2, node: rightIngNode, depth: current.depth + 1})
				}

				current.node.combinations = []Recipe{{
					ingredient1: leftIngNode,
					ingredient2: rightIngNode,
				}}
			}

			resultChan <- rootNode
		}(pair)
	}

	// Close collector
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for tree := range resultChan {
		mu.Lock()
		if maxPathsToReturn <= 0 || len(allFinalTargetTrees) < maxPathsToReturn {
			allFinalTargetTrees = append(allFinalTargetTrees, tree)
		}
		mu.Unlock()
	}

	return allFinalTargetTrees
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
