package main

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

type PathResult struct {
	path             *Node
	score            int
	desc             string
	actualTargetLeaf *Node
	pathSignature    string // Add a unique signature to identify duplicate paths
}

// findMultipleBidirectionalPaths melakukan pencarian dua arah secara konkuren
// dan mengembalikan n jalur valid pertama dari root ke elemen dasar.
func findMultipleBidirectionalPaths(tree *Tree, numPaths int) []*Node {
	if tree == nil || tree.root == nil {
		fmt.Println("Tree is empty, cannot perform search.")
		return nil
	}

	if tree.root.isCycleNode {
		fmt.Println("Root is a cycle node, cannot perform search.")
		return nil
	}

	fmt.Printf("\n--- Starting Multi-Threaded Bidirectional Search for %d Paths ---\n", numPaths)

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

	resultChan := make(chan PathResult, numPaths)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	var foundPaths []*Node
	var numFoundPaths int

	// Map to track path signatures we've already added
	pathSignatures := make(map[string]bool)

	// Start concurrent searches from each base leaf
	for i, singleBaseLeaf := range baseLeaves {
		wg.Add(1)
		go func(leafIndex int, currentTargetLeaf *Node) {
			defer wg.Done()

			resultsMutex.Lock()
			if numFoundPaths >= numPaths {
				resultsMutex.Unlock()
				return
			}
			resultsMutex.Unlock()

			path, score, desc, actualLeafFound := bidirectionalSearchFromLeaf(tree.root, []*Node{currentTargetLeaf})

			if path != nil {
				// Generate a unique signature for this path
				pathSignature := generatePathSignature(path)

				resultChan <- PathResult{
					path:             path,
					score:            score,
					desc:             desc,
					actualTargetLeaf: actualLeafFound,
					pathSignature:    pathSignature,
				}
			}
		}(i, singleBaseLeaf)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		resultsMutex.Lock()

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
// This function traverses the entire path and creates a string representing the sequence of elements
func generatePathSignature(node *Node) string {
	if node == nil {
		return ""
	}

	// Use a breadth-first traversal to collect all nodes in the path
	var result []string
	queue := list.New()
	visited := make(map[*Node]bool)

	queue.PushBack(node)
	visited[node] = true

	for queue.Len() > 0 {
		current := queue.Front().Value.(*Node)
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

	return strings.Join(result, "|")
}

// bidirectionalSearchFromLeaf melakukan pencarian dua arah antara root dan sekumpulan targetBaseLeaves.
// Fungsi ini menemukan jalur ke salah satu dari targetBaseLeaves tersebut.
// Mengembalikan pohon jalur, skor, deskripsi, dan node leaf aktual tempat jalur ditemukan.
func bidirectionalSearchFromLeaf(root *Node, targetBaseLeaves []*Node) (*Node, int, string, *Node) {
	// Struct MeetingPoint didefinisikan secara lokal untuk mengakomodasi field tambahan
	// tanpa mengubah definisi global jika ada.
	type MeetingPoint struct {
		node             *Node
		forwardDepth     int
		backwardDepth    int
		actualTargetLeaf *Node // Leaf spesifik yang terhubung dengan meeting point ini
	}

	if root == nil {
		return nil, -1, "", nil
	}
	// Periksa apakah targetBaseLeaves kosong atau semua elemennya tidak valid
	if len(targetBaseLeaves) == 0 {
		// fmt.Println("No target base leaves provided to bidirectionalSearchFromLeaf.") // Opsional
		return nil, -1, "", nil
	}

	q_f := list.New()
	visited_f := make(map[*Node]*Node)
	q_f.PushBack(root)
	visited_f[root] = nil
	forwardDepth := make(map[*Node]int)
	forwardDepth[root] = 0

	q_b := list.New()
	visited_b := make(map[*Node]*Node)
	backwardDepth := make(map[*Node]int)
	baseLeafSource := make(map[*Node]*Node) // Melacak asal base leaf untuk jalur mundur

	// Inisialisasi pencarian mundur dari semua targetBaseLeaves
	validTargetsFound := false
	for _, leaf := range targetBaseLeaves {
		if leaf == nil || leaf.isCycleNode { // Lewati leaf yang tidak valid
			continue
		}
		q_b.PushBack(leaf)
		visited_b[leaf] = nil
		backwardDepth[leaf] = 0
		baseLeafSource[leaf] = leaf // Leaf ini adalah sumbernya sendiri
		validTargetsFound = true
	}

	if !validTargetsFound { // Jika tidak ada target valid setelah filter
		// fmt.Println("No valid target base leaves to start backward search from in bidirectionalSearchFromLeaf.") // Opsional
		return nil, -1, "", nil
	}

	var meetingPoints []MeetingPoint

	// Komentari atau sesuaikan info debug karena baseLeaf sekarang adalah slice
	// fmt.Printf("Starting bi-directional search: root=%s to one of target leaves\n", root.element)

	for q_f.Len() > 0 && q_b.Len() > 0 {
		// Proses satu level pencarian maju
		currLevelSize_f := q_f.Len()
		for i := 0; i < currLevelSize_f; i++ {
			frontElement_f := q_f.Front()
			curr_f_instance := frontElement_f.Value.(*Node)
			q_f.Remove(frontElement_f)

			if curr_f_instance.isCycleNode {
				continue
			}

			if backDepth, found := backwardDepth[curr_f_instance]; found {
				// Pastikan baseLeafSource[curr_f_instance] ada (seharusnya ada jika found)
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
				children := []*Node{recipe.ingredient1, recipe.ingredient2}
				for _, child_instance := range children {
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
									backwardDepth:    backDepth, // Menggunakan backDepth dari lookup, bukan b_found (boolean)
									actualTargetLeaf: actualLeaf,
								})
							}
						}
					}
				}
			}
		}

		// Proses satu level pencarian mundur
		currLevelSize_b := q_b.Len()
		for i := 0; i < currLevelSize_b; i++ {
			frontElement_b := q_b.Front()
			curr_b_instance := frontElement_b.Value.(*Node)
			q_b.Remove(frontElement_b)

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
			if parent_instance == nil || parent_instance.isCycleNode {
				continue
			}
			if _, v_found := visited_b[parent_instance]; !v_found {
				q_b.PushBack(parent_instance)
				visited_b[parent_instance] = curr_b_instance
				backwardDepth[parent_instance] = backwardDepth[curr_b_instance] + 1
				if sourceLeaf, ok := baseLeafSource[curr_b_instance]; ok { // Propagasi sumber
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
		// Komentari atau sesuaikan info debug karena baseLeaf sekarang adalah slice
		// fmt.Printf("No meeting point found between %s and target leaves\n", root.element)
		return nil, -1, "", nil
	}

	var bestMeetingPointData MeetingPoint
	foundBest := false
	bestTotalDepth := -1

	for _, mp := range meetingPoints {
		totalDepth := mp.forwardDepth + mp.backwardDepth
		if !foundBest || totalDepth < bestTotalDepth {
			bestTotalDepth = totalDepth
			bestMeetingPointData = mp // mp adalah struct, bukan pointer, jadi ini adalah copy
			foundBest = true
		} else if totalDepth == bestTotalDepth {
			if bestMeetingPointData.actualTargetLeaf != nil && mp.actualTargetLeaf != nil && // Pastikan tidak nil
				mp.backwardDepth < bestMeetingPointData.backwardDepth {
				bestMeetingPointData = mp
			}
		}
	}

	if !foundBest || bestMeetingPointData.actualTargetLeaf == nil { // Periksa juga actualTargetLeaf
		return nil, -1, "", nil
	}

	pathDesc := fmt.Sprintf("Path to %s (total depth: %d)",
		bestMeetingPointData.actualTargetLeaf.element, bestTotalDepth)

	// Akses ke recipeData tetap seperti di kode asli Anda, diasumsikan sebagai variabel global/package-level.
	// Tidak ada perubahan di sini sesuai permintaan "jangan buat perubahan selain..."
	recipesMapConverted := map[string][][]string(recipeData.Recipes)
	resultTree := constructShortestPathTree(bestMeetingPointData.node, visited_f, visited_b, recipesMapConverted)

	return resultTree, bestTotalDepth, pathDesc, bestMeetingPointData.actualTargetLeaf
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
