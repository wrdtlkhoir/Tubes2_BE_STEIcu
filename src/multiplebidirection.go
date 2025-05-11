package main

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

type SimpleOutputData struct {
	Elements []string              `json:"elements"` // List of all element names
	Recipes  map[string][][]string `json:"recipes"`  // Map of element name to its recipes
}

var recipeDatas SimpleOutputData

type PathResult struct {
	path             *Nodebidir
	score            int
	desc             string
	actualTargetLeaf *Nodebidir
	pathSignature    string // Add a unique signature to identify duplicate paths
}

func isPathCyclic(node *Nodebidir) bool {
	if node == nil {
		return false
	}

	// Use BFS to traverse the entire path tree
	queue := list.New()
	visited := make(map[*Nodebidir]bool)

	queue.PushBack(node)
	visited[node] = true

	for queue.Len() > 0 {
		current := queue.Front().Value.(*Nodebidir)
		queue.Remove(queue.Front())

		// Check if the current node is cyclic
		if current.isCycleNode {
			return true
		}

		// Add all combinations (children) to the queue
		for _, recipe := range current.combinations {
			// Check first ingredient
			if recipe.ingredient1 != nil && !visited[recipe.ingredient1] {
				if recipe.ingredient1.isCycleNode {
					return true
				}
				queue.PushBack(recipe.ingredient1)
				visited[recipe.ingredient1] = true
			}

			// Check second ingredient
			if recipe.ingredient2 != nil && !visited[recipe.ingredient2] {
				if recipe.ingredient2.isCycleNode {
					return true
				}
				queue.PushBack(recipe.ingredient2)
				visited[recipe.ingredient2] = true
			}
		}
	}

	return false
}

// findMultipleBidirectionalPaths performs concurrent bidirectional search
// and returns the first n valid paths from root to base elements.
func findMultipleBidirectionalPaths(tree *Treebidir, numPaths int) []*Nodebidir {
	if tree == nil || tree.root == nil {
		fmt.Println("Tree is empty, cannot perform search.")
		return nil
	}

	if tree.root.isCycleNode {
		fmt.Println("Root is a cycle node, cannot perform search.")
		return nil
	}

	fmt.Printf("\n--- Starting Multi-Threaded Bidirectional Search for %d Paths ---\n", numPaths)

	baseLeaves := findBaseLeaves(tree.root, []*Nodebidir{})
	if len(baseLeaves) == 0 {
		fmt.Println("No base leaves found in the tree, cannot perform backward search.")
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
		fmt.Println("No valid non-cyclic base leaves found in the tree, cannot perform backward search.")
		return nil
	}

	fmt.Printf("Found %d valid non-cyclic base leaves for backward search: ", len(validLeaves))
	for i, leaf := range validLeaves {
		fmt.Print(leaf.element)
		if i < len(validLeaves)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println()

	resultChan := make(chan PathResult, numPaths*2) // Increase buffer size to avoid potential deadlocks
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	var foundPaths []*Nodebidir
	var numFoundPaths int

	// Map to track path signatures we've already added
	pathSignatures := make(map[string]bool)

	// Start concurrent searches ONLY from valid non-cyclic base leaves
	for i, singleBaseLeaf := range validLeaves {
		wg.Add(1)
		go func(leafIndex int, currentTargetLeaf *Nodebidir) {
			defer wg.Done()

			// Check if we should continue searching based on paths found so far
			resultsMutex.Lock()
			shouldContinue := numFoundPaths < numPaths
			resultsMutex.Unlock()

			if !shouldContinue {
				return
			}

			// We already filtered out cyclic leaves, but double-check just in case
			if currentTargetLeaf.isCycleNode {
				fmt.Printf("Skipping cyclic leaf: %s (safety check)\n", currentTargetLeaf.element)
				return
			}

			path, score, desc, actualLeafFound := bidirectionalSearchFromLeaf(tree.root, []*Nodebidir{currentTargetLeaf})

			// Only proceed if we found a path and it doesn't contain any cyclic nodes
			if path != nil && !isPathCyclic(path) {
				// Generate a unique signature for this path
				pathSignature := generatePathSignature(path)

				resultChan <- PathResult{
					path:             path,
					score:            score,
					desc:             desc,
					actualTargetLeaf: actualLeafFound,
					pathSignature:    pathSignature,
				}
			} else if path != nil {
				// Debug output for paths containing cyclic nodes
				fmt.Printf("Found path with cyclic nodes: %s (discarded)\n", desc)
			}
		}(i, singleBaseLeaf)
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

	fmt.Printf("--- Completed Multi-Threaded Search, found %d paths ---\n", len(foundPaths))
	return foundPaths
}

// generatePathSignature creates a unique string identifier for a path
func generatePathSignature(node *Nodebidir) string {
	if node == nil {
		return ""
	}

	// Use a breadth-first traversal to collect all nodes in the path
	var result []string
	queue := list.New()
	visited := make(map[*Nodebidir]bool)

	queue.PushBack(node)
	visited[node] = true

	for queue.Len() > 0 {
		current := queue.Front().Value.(*Nodebidir)
		queue.Remove(queue.Front())

		// Add this node's element to the signature
		result = append(result, current.element)

		// Add all combinations (children)
		for _, recipe := range current.combinations {
			if recipe.ingredient1 != nil && !visited[recipe.ingredient1] {
				queue.PushBack(recipe.ingredient1)
				visited[recipe.ingredient1] = true
			}
			if recipe.ingredient2 != nil && !visited[recipe.ingredient2] {
				queue.PushBack(recipe.ingredient2)
				visited[recipe.ingredient2] = true
			}
		}
	}

	// Sort the elements to ensure the same path structure always generates the same signature
	return strings.Join(result, "|")
}

func bidirectionalSearchFromLeaf(root *Nodebidir, targetBaseLeaves []*Nodebidir) (*Nodebidir, int, string, *Nodebidir) {
	// Struct to hold meeting point data
	type MeetingPoint struct {
		node             *Nodebidir
		forwardDepth     int
		backwardDepth    int
		actualTargetLeaf *Nodebidir
	}

	if root == nil {
		return nil, -1, "", nil
	}
	// Check if targetBaseLeaves is empty or all elements invalid
	if len(targetBaseLeaves) == 0 {
		return nil, -1, "", nil
	}

	// Skip if root is a cyclic node
	if root.isCycleNode {
		return nil, -1, "", nil
	}

	q_f := list.New()
	visited_f := make(map[*Nodebidir]*Nodebidir)
	q_f.PushBack(root)
	visited_f[root] = nil
	forwardDepth := make(map[*Nodebidir]int)
	forwardDepth[root] = 0

	q_b := list.New()
	visited_b := make(map[*Nodebidir]*Nodebidir)
	backwardDepth := make(map[*Nodebidir]int)
	baseLeafSource := make(map[*Nodebidir]*Nodebidir) // Track source base leaf for backward path

	// Initialize backward search from all targetBaseLeaves
	validTargetsFound := false
	for _, leaf := range targetBaseLeaves {
		if leaf == nil || leaf.isCycleNode { // Skip invalid leaves
			continue
		}
		q_b.PushBack(leaf)
		visited_b[leaf] = nil
		backwardDepth[leaf] = 0
		baseLeafSource[leaf] = leaf // This leaf is its own source
		validTargetsFound = true
	}

	if !validTargetsFound { // If no valid targets after filtering
		return nil, -1, "", nil
	}

	var meetingPoints []MeetingPoint

	for q_f.Len() > 0 && q_b.Len() > 0 {
		// Process one level of forward search
		currLevelSize_f := q_f.Len()
		for i := 0; i < currLevelSize_f; i++ {
			frontElement_f := q_f.Front()
			curr_f_instance := frontElement_f.Value.(*Nodebidir)
			q_f.Remove(frontElement_f)

			// Skip cyclic nodes entirely
			if curr_f_instance.isCycleNode {
				continue
			}

			if backDepth, found := backwardDepth[curr_f_instance]; found {
				// Ensure baseLeafSource[curr_f_instance] exists (should exist if found)
				if actualLeaf, ok := baseLeafSource[curr_f_instance]; ok {
					meetingPoints = append(meetingPoints, MeetingPoint{
						node:             curr_f_instance,
						forwardDepth:     forwardDepth[curr_f_instance],
						backwardDepth:    backDepth,
						actualTargetLeaf: actualLeaf,
					})
				}
			}

			for _, recipe := range curr_f_instance.combinations {
				children := []*Nodebidir{recipe.ingredient1, recipe.ingredient2}
				for _, child_instance := range children {
					// Skip nil or cyclic child nodes entirely
					if child_instance == nil || child_instance.isCycleNode {
						continue
					}
					if _, v_found := visited_f[child_instance]; !v_found {
						q_f.PushBack(child_instance)
						visited_f[child_instance] = curr_f_instance
						forwardDepth[child_instance] = forwardDepth[curr_f_instance] + 1
						if backDepth, b_found := backwardDepth[child_instance]; b_found {
							if actualLeaf, ok := baseLeafSource[child_instance]; ok {
								meetingPoints = append(meetingPoints, MeetingPoint{
									node:             child_instance,
									forwardDepth:     forwardDepth[child_instance],
									backwardDepth:    backDepth,
									actualTargetLeaf: actualLeaf,
								})
							}
						}
					}
				}
			}
		}

		// Process one level of backward search
		currLevelSize_b := q_b.Len()
		for i := 0; i < currLevelSize_b; i++ {
			frontElement_b := q_b.Front()
			curr_b_instance := frontElement_b.Value.(*Nodebidir)
			q_b.Remove(frontElement_b)

			// Skip cyclic nodes entirely
			if curr_b_instance.isCycleNode {
				continue
			}

			if fwdDepth, found := forwardDepth[curr_b_instance]; found {
				if actualLeaf, ok := baseLeafSource[curr_b_instance]; ok {
					meetingPoints = append(meetingPoints, MeetingPoint{
						node:             curr_b_instance,
						forwardDepth:     fwdDepth,
						backwardDepth:    backwardDepth[curr_b_instance],
						actualTargetLeaf: actualLeaf,
					})
				}
			}

			parent_instance := curr_b_instance.parent
			// Skip nil or cyclic parent nodes entirely
			if parent_instance == nil || parent_instance.isCycleNode {
				continue
			}
			if _, v_found := visited_b[parent_instance]; !v_found {
				q_b.PushBack(parent_instance)
				visited_b[parent_instance] = curr_b_instance
				backwardDepth[parent_instance] = backwardDepth[curr_b_instance] + 1
				if sourceLeaf, ok := baseLeafSource[curr_b_instance]; ok { // Propagate source
					baseLeafSource[parent_instance] = sourceLeaf
					if fwdDepth, f_found := forwardDepth[parent_instance]; f_found {
						meetingPoints = append(meetingPoints, MeetingPoint{
							node:             parent_instance,
							forwardDepth:     fwdDepth,
							backwardDepth:    backwardDepth[parent_instance],
							actualTargetLeaf: sourceLeaf,
						})
					}
				}
			}
		}

		if len(meetingPoints) > 0 {
			break
		}
	}

	if len(meetingPoints) == 0 {
		return nil, -1, "", nil
	}

	var bestMeetingPointData MeetingPoint
	foundBest := false
	bestTotalDepth := -1

	for _, mp := range meetingPoints {
		totalDepth := mp.forwardDepth + mp.backwardDepth
		if !foundBest || totalDepth < bestTotalDepth {
			bestTotalDepth = totalDepth
			bestMeetingPointData = mp // mp is a struct, not a pointer, so this is a copy
			foundBest = true
		} else if totalDepth == bestTotalDepth {
			if bestMeetingPointData.actualTargetLeaf != nil && mp.actualTargetLeaf != nil && // Ensure not nil
				mp.backwardDepth < bestMeetingPointData.backwardDepth {
				bestMeetingPointData = mp
			}
		}
	}

	if !foundBest || bestMeetingPointData.actualTargetLeaf == nil { // Also check actualTargetLeaf
		return nil, -1, "", nil
	}

	pathDesc := fmt.Sprintf("Path to %s (total depth: %d)",
		bestMeetingPointData.actualTargetLeaf.element, bestTotalDepth)

	// Construct the path tree
	recipesMapConverted := map[string][][]string(recipeDatas.Recipes)
	resultTree := constructShortestPathTree(bestMeetingPointData.node, visited_f, visited_b, recipesMapConverted)

	// Final check to ensure no cyclic nodes in the constructed tree
	if isPathCyclic(resultTree) {
		fmt.Printf("Rejecting path with cyclic nodes: %s\n", pathDesc)
		return nil, -1, "", nil
	}

	return resultTree, bestTotalDepth, pathDesc, bestMeetingPointData.actualTargetLeaf
}

// Replacement for the main function to demonstrate the multi-path implementation
func mainWithMultiplePaths(recipeData map[string][][]string, numPaths int) {
	targetElement := "Human"

	// Build tree from data
	fullTree := buildTreeBFS(targetElement, recipeData)

	// Print the full tree
	fmt.Println("\nFull Recipe Derivation Tree:")
	printTreeBidir(fullTree)

	// Perform multi-path bidirectional search
	foundPaths := findMultipleBidirectionalPaths(fullTree, numPaths)

	fmt.Printf("\nFound %d paths:\n", len(foundPaths))
	for i, path := range foundPaths {
		fmt.Printf("\nPath %d:\n", i+1)
		printShortestPathTree(path, "", true)
	}

	fmt.Println("\n" + strings.Repeat("=", 40)) // Separator
}
