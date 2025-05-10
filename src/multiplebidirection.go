package main

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

type PathResult struct {
	path  *Node  // Root of the path tree
	score int    // Total depth/score of the path
	desc  string // Description of the path
}

// findMultipleBidirectionalPaths performs concurrent bidirectional searches
// and returns the first n valid paths from root to base elements.
func findMultipleBidirectionalPaths(tree *Tree, numPaths int) []*Node {
	if tree == nil || tree.root == nil {
		fmt.Println("Tree is empty, cannot perform search.")
		return nil
	}

	// Skip if the root itself is a cycle node
	if tree.root.isCycleNode {
		fmt.Println("Root is a cycle node, cannot perform search.")
		return nil
	}

	fmt.Printf("\n--- Starting Multi-Threaded Bidirectional Search for %d Paths ---\n", numPaths)

	// Find base leaves for backward search
	baseLeaves := findBaseLeaves(tree.root, []*Node{})
	if len(baseLeaves) == 0 {
		fmt.Println("No base leaves found in the tree, cannot perform backward search.")
		return nil
	}

	fmt.Printf("Found %d base leaves for backward search: ", len(baseLeaves))
	for i, leaf := range baseLeaves {
		fmt.Print(leaf.element)
		if i < len(baseLeaves)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println()

	// Channel to collect results
	resultChan := make(chan PathResult, numPaths)

	// WaitGroup to track completion of all searches
	var wg sync.WaitGroup

	// Mutex to protect shared data structures
	var resultsMutex sync.Mutex
	var foundPaths []*Node
	var numFoundPaths int

	// Start concurrent searches from each base leaf
	for i, baseLeaf := range baseLeaves {
		wg.Add(1)
		go func(leafIndex int, leaf *Node) {
			defer wg.Done()

			// Check if we already have enough paths before starting this search
			resultsMutex.Lock()
			if numFoundPaths >= numPaths {
				resultsMutex.Unlock()
				return
			}
			resultsMutex.Unlock()

			// Run bidirectional search for this base leaf
			path, score, desc := bidirectionalSearchFromLeaf(tree.root, leaf)

			if path != nil {
				// Send result to channel
				resultChan <- PathResult{
					path:  path,
					score: score,
					desc:  desc,
				}
			}
		}(i, baseLeaf)
	}

	// Goroutine to close result channel when all searches are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results up to numPaths
	for result := range resultChan {
		resultsMutex.Lock()
		if numFoundPaths < numPaths {
			foundPaths = append(foundPaths, result.path)
			numFoundPaths++
			fmt.Printf("Found path %d/%d: %s\n", numFoundPaths, numPaths, result.desc)
		}
		resultsMutex.Unlock()

		// Break if we have enough paths
		if numFoundPaths >= numPaths {
			break
		}
	}

	fmt.Printf("--- Completed Multi-Threaded Search, found %d paths ---\n", len(foundPaths))
	return foundPaths
}

// bidirectionalSearchFromLeaf performs bidirectional search between root and a specific leaf
func bidirectionalSearchFromLeaf(root *Node, baseLeaf *Node) (*Node, int, string) {
	// Forward search (starting from the root)
	q_f := list.New()
	visited_f := make(map[*Node]*Node) // node instance -> parent node instance

	q_f.PushBack(root)
	visited_f[root] = nil

	// Backward search (starting from the specified base leaf)
	q_b := list.New()
	visited_b := make(map[*Node]*Node) // node instance -> parent node instance

	q_b.PushBack(baseLeaf)
	visited_b[baseLeaf] = nil

	// Track depth in both directions
	forwardDepth := make(map[*Node]int)
	backwardDepth := make(map[*Node]int)

	forwardDepth[root] = 0
	backwardDepth[baseLeaf] = 0

	// Meeting points tracking
	var meetingPoints []MeetingPoint

	// Debug info
	fmt.Printf("Starting bi-directional search: root=%s to leaf=%s\n",
		root.element, baseLeaf.element)

	// Perform Bi-BFS
	for q_f.Len() > 0 && q_b.Len() > 0 {
		// Process one level of forward search
		currLevelSize_f := q_f.Len()
		for i := 0; i < currLevelSize_f; i++ {
			frontElement_f := q_f.Front()
			curr_f_instance := frontElement_f.Value.(*Node)
			q_f.Remove(frontElement_f)

			// Skip cycle nodes in forward direction
			if curr_f_instance.isCycleNode {
				continue
			}

			// Check for collision with backward search
			if backDepth, found := backwardDepth[curr_f_instance]; found {
				meetingPoints = append(meetingPoints, MeetingPoint{
					node:          curr_f_instance,
					forwardDepth:  forwardDepth[curr_f_instance],
					backwardDepth: backDepth,
					baseLeaf:      baseLeaf,
				})
				// Found a meeting point, but continue searching
			}

			// Expand forward: Add ingredient *Node instances* from combinations
			for _, recipe := range curr_f_instance.combinations {
				children := []*Node{recipe.ingredient1, recipe.ingredient2}

				for _, child_instance := range children {
					if child_instance == nil || child_instance.isCycleNode {
						continue
					}

					// Skip if already visited in forward direction
					if _, v_found := visited_f[child_instance]; !v_found {
						q_f.PushBack(child_instance)
						visited_f[child_instance] = curr_f_instance
						forwardDepth[child_instance] = forwardDepth[curr_f_instance] + 1

						// Check for collision immediately
						if backDepth, b_found := backwardDepth[child_instance]; b_found {
							meetingPoints = append(meetingPoints, MeetingPoint{
								node:          child_instance,
								forwardDepth:  forwardDepth[child_instance],
								backwardDepth: backDepth,
								baseLeaf:      baseLeaf,
							})
						}
					}
				}
			}
		}

		// Process one level of backward search
		currLevelSize_b := q_b.Len()
		for i := 0; i < currLevelSize_b; i++ {
			frontElement_b := q_b.Front()
			curr_b_instance := frontElement_b.Value.(*Node)
			q_b.Remove(frontElement_b)

			// Skip cycle nodes in backward direction
			if curr_b_instance.isCycleNode {
				continue
			}

			// Check for collision with forward search
			if fwdDepth, found := forwardDepth[curr_b_instance]; found {
				meetingPoints = append(meetingPoints, MeetingPoint{
					node:          curr_b_instance,
					forwardDepth:  fwdDepth,
					backwardDepth: backwardDepth[curr_b_instance],
					baseLeaf:      baseLeaf,
				})
				// Found a meeting point, but continue searching
			}

			// Expand backward: Move up to the parent *Node instance*
			parent_instance := curr_b_instance.parent

			if parent_instance == nil || parent_instance.isCycleNode {
				continue
			}

			// Check if this parent node instance has been visited in backward search
			if _, v_found := visited_b[parent_instance]; !v_found {
				q_b.PushBack(parent_instance)
				visited_b[parent_instance] = curr_b_instance
				backwardDepth[parent_instance] = backwardDepth[curr_b_instance] + 1

				// Check for collision immediately
				if fwdDepth, f_found := forwardDepth[parent_instance]; f_found {
					meetingPoints = append(meetingPoints, MeetingPoint{
						node:          parent_instance,
						forwardDepth:  fwdDepth,
						backwardDepth: backwardDepth[parent_instance],
						baseLeaf:      baseLeaf,
					})
				}
			}
		}

		// If we found any meeting point, we can break early to get the first path
		if len(meetingPoints) > 0 {
			break
		}
	}

	// If no meeting point found, return nil
	if len(meetingPoints) == 0 {
		fmt.Printf("No meeting point found between %s and %s\n",
			root.element, baseLeaf.element)
		return nil, -1, ""
	}

	// Find the meeting point with the shortest total path
	var bestMeetingPoint *MeetingPoint
	bestTotalDepth := -1

	for i := range meetingPoints {
		totalDepth := meetingPoints[i].forwardDepth + meetingPoints[i].backwardDepth

		if bestTotalDepth == -1 || totalDepth < bestTotalDepth {
			bestTotalDepth = totalDepth
			bestMeetingPoint = &meetingPoints[i]
		}
	}

	if bestMeetingPoint == nil {
		return nil, -1, ""
	}

	// Create a description of the path
	pathDesc := fmt.Sprintf("Path to %s (total depth: %d)",
		baseLeaf.element, bestTotalDepth)

	// Construct the shortest path tree
	recipesMapConverted := map[string][][]string(recipeData.Recipes)
	resultTree := constructShortestPathTree(bestMeetingPoint.node, visited_f, visited_b, recipesMapConverted)

	return resultTree, bestTotalDepth, pathDesc
}

// Replacement for the main function to demonstrate the multi-path implementation
func mainWithMultiplePaths(recipeData map[string][][]string, numPaths int) {
	targetElement := "Swamp"

	// Build tree from data
	fullTree := buildTreeBFS(targetElement, recipeData)

	// Print the full tree
	fmt.Println("\nFull Recipe Derivation Tree:")
	printTree(fullTree)

	// Perform multi-path bidirectional search
	foundPaths := findMultipleBidirectionalPaths(fullTree, numPaths)

	fmt.Printf("\nFound %d paths:\n", len(foundPaths))
	for i, path := range foundPaths {
		fmt.Printf("\nPath %d:\n", i+1)
		printShortestPathTree(path, "", true)
	}

	fmt.Println("\n" + strings.Repeat("=", 40)) // Separator
}
