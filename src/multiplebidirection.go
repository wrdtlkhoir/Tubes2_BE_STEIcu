package main

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

type SimpleOutputData struct {
	Elements []string              `json:"elements"`
	Recipes  map[string][][]string `json:"recipes"`
}

var recipeDatas SimpleOutputData

type PathResult struct {
	path             *Nodebidir
	score            int
	desc             string
	actualTargetLeaf *Nodebidir
	pathSignature    string
	exploredNodes    int
}

func isPathCyclic(node *Nodebidir) bool {
	if node == nil {
		return false
	}

	queue := list.New()
	visited := make(map[*Nodebidir]bool)

	queue.PushBack(node)
	visited[node] = true

	for queue.Len() > 0 {
		current := queue.Front().Value.(*Nodebidir)
		queue.Remove(queue.Front())

		if current.isCycleNode {
			return true
		}

		for _, recipe := range current.combinations {
			if recipe.ingredient1 != nil && !visited[recipe.ingredient1] {
				if recipe.ingredient1.isCycleNode {
					return true
				}
				queue.PushBack(recipe.ingredient1)
				visited[recipe.ingredient1] = true
			}

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

func findMultipleBidirectionalPaths(tree *Treebidir, numPaths int) ([]*Nodebidir, int) {
	totalExploredNodesOverall := 0 // Akumulator untuk semua node yang dieksplor di semua pencarian

	if tree == nil || tree.root == nil {
		fmt.Println("Tree is empty")
		return nil, totalExploredNodesOverall
	}
	if tree.root.isCycleNode {
		fmt.Println("Root is a cycle node")
		return nil, totalExploredNodesOverall
	}

	baseLeaves := findBaseLeaves(tree.root, []*Nodebidir{})
	if len(baseLeaves) == 0 {
		fmt.Println("No base leaves found")
		return nil, totalExploredNodesOverall
	}

	var validLeaves []*Nodebidir
	for _, leaf := range baseLeaves {
		if !leaf.isCycleNode {
			validLeaves = append(validLeaves, leaf)
		}
	}

	if len(validLeaves) == 0 {
		fmt.Println("No valid (non-cyclic) base leaves found")
		return nil, totalExploredNodesOverall
	}

	// Perbesar buffer channel jika banyak leaf dan numPaths kecil,
	// agar goroutine tidak block saat mengirim PathResult jika resultChan penuh
	// sebelum consumer sempat mengambil. Max(numPaths*2, len(validLeaves)) bisa jadi pilihan.
	resultChan := make(chan PathResult, len(validLeaves))
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex // Melindungi foundPaths, numFoundPaths, pathSignatures
	var foundPaths []*Nodebidir
	var numFoundPaths int

	pathSignatures := make(map[string]bool)

	for i, singleBaseLeaf := range validLeaves {
		// Jika ingin membatasi jumlah goroutine yang berjalan sekaligus,
		// bisa menggunakan semaphore channel di sini.
		// Untuk sekarang, kita jalankan untuk semua validLeaves.

		wg.Add(1)
		go func(leafIndex int, currentTargetLeaf *Nodebidir) {
			defer wg.Done()

			// Cek apakah kita masih perlu mencari jalur baru
			resultsMutex.Lock()
			shouldSearch := numFoundPaths < numPaths
			resultsMutex.Unlock()

			if !shouldSearch {
				// Kirim hasil dengan exploredNodes = 0 jika tidak melakukan search
				resultChan <- PathResult{exploredNodes: 0}
				return
			}

			if currentTargetLeaf.isCycleNode { // Sebenarnya sudah difilter di validLeaves
				resultChan <- PathResult{exploredNodes: 0}
				return
			}

			// Jalankan pencarian untuk leaf ini
			path, score, desc, actualLeafFound, exploredForThisSearch := bidirectionalSearchFromLeaf(tree.root, []*Nodebidir{currentTargetLeaf})

			pathSignature := ""
			if path != nil && !isPathCyclic(path) {
				pathSignature = generatePathSignature(path)
			} else {
				path = nil // Pastikan path nil jika siklik atau tidak ditemukan
			}

			resultChan <- PathResult{
				path:             path,
				score:            score,
				desc:             desc,
				actualTargetLeaf: actualLeafFound,
				pathSignature:    pathSignature,
				exploredNodes:    exploredForThisSearch, // Sertakan jumlah node dieksplor
			}
		}(i, singleBaseLeaf)
	}

	// Goroutine untuk menutup channel setelah semua pencarian selesai
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Kumpulkan hasil dari channel
	for result := range resultChan {
		totalExploredNodesOverall += result.exploredNodes // Akumulasi total node dieksplor dari setiap pencarian

		if result.path == nil || result.pathSignature == "" { // Jika path nil atau signature kosong (karena path nil/siklik)
			continue
		}

		resultsMutex.Lock()
		// Cek lagi karena kondisi numFoundPaths bisa berubah oleh goroutine lain yang masuk ke channel lebih dulu
		if numFoundPaths < numPaths && !pathSignatures[result.pathSignature] {
			// isPathCyclic sudah dicek di goroutine, tapi bisa dicek lagi di sini jika paranoid
			// if isPathCyclic(result.path) {
			// 	resultsMutex.Unlock()
			// 	continue
			// }
			foundPaths = append(foundPaths, result.path)
			numFoundPaths++
			pathSignatures[result.pathSignature] = true
			// fmt.Printf("Path %d found (score %d), target leaf %s. Explored by this search: %d\n", numFoundPaths, result.score, result.actualTargetLeaf.element, result.exploredNodes)
		}
		resultsMutex.Unlock()

		// Jika sudah cukup path, kita bisa berhenti lebih awal,
		// tapi kita tetap perlu mengosongkan channel untuk menghitung semua exploredNodes
		// dan agar goroutine wg.Wait() tidak deadlock.
		// Loop ini akan berhenti ketika channel ditutup.
	}

	return foundPaths, totalExploredNodesOverall
}

func generatePathSignature(node *Nodebidir) string {
	if node == nil {
		return ""
	}
	var result []string
	queue := list.New()
	visited := make(map[*Nodebidir]bool)

	queue.PushBack(node)
	visited[node] = true

	for queue.Len() > 0 {
		current := queue.Front().Value.(*Nodebidir)
		queue.Remove(queue.Front())
		result = append(result, current.element)
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

func bidirectionalSearchFromLeaf(root *Nodebidir, targetBaseLeaves []*Nodebidir) (*Nodebidir, int, string, *Nodebidir, int) {
	exploredNodesCount := 0 // Inisialisasi penghitung

	type MeetingPoint struct {
		node             *Nodebidir
		forwardDepth     int
		backwardDepth    int
		actualTargetLeaf *Nodebidir
	}

	if root == nil {
		return nil, -1, "", nil, exploredNodesCount
	}
	if len(targetBaseLeaves) == 0 {
		return nil, -1, "", nil, exploredNodesCount
	}
	if root.isCycleNode { // Cek root node siklus
		return nil, -1, "", nil, exploredNodesCount
	}

	q_f := list.New()
	visited_f := make(map[*Nodebidir]*Nodebidir)
	q_f.PushBack(root)
	visited_f[root] = nil
	forwardDepth := make(map[*Nodebidir]int)
	forwardDepth[root] = 0
	// root akan dihitung saat di-pop

	q_b := list.New()
	visited_b := make(map[*Nodebidir]*Nodebidir)
	backwardDepth := make(map[*Nodebidir]int)
	baseLeafSource := make(map[*Nodebidir]*Nodebidir)

	validTargetsFound := false
	for _, leaf := range targetBaseLeaves {
		if leaf == nil || leaf.isCycleNode {
			continue
		}
		q_b.PushBack(leaf)
		visited_b[leaf] = nil
		backwardDepth[leaf] = 0
		baseLeafSource[leaf] = leaf // Melacak leaf asli untuk setiap node di pencarian mundur
		validTargetsFound = true
		// leaf akan dihitung saat di-pop
	}

	if !validTargetsFound {
		return nil, -1, "", nil, exploredNodesCount
	}

	var meetingPoints []MeetingPoint

	for q_f.Len() > 0 && q_b.Len() > 0 {
		// Forward search step
		if q_f.Len() > 0 { // Perlu dicek lagi karena q_b mungkin jadi kosong di iterasi sebelumnya
			frontElement_f := q_f.Front()
			curr_f_instance := frontElement_f.Value.(*Nodebidir)
			q_f.Remove(frontElement_f)
			exploredNodesCount++ // Hitung

			if curr_f_instance.isCycleNode {
				continue
			}

			if backDepth, found := backwardDepth[curr_f_instance]; found {
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

		// Backward search step
		if q_b.Len() > 0 { // Perlu dicek lagi karena q_f mungkin jadi kosong
			frontElement_b := q_b.Front()
			curr_b_instance := frontElement_b.Value.(*Nodebidir)
			q_b.Remove(frontElement_b)
			exploredNodesCount++ // Hitung

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
				if sourceLeaf, ok := baseLeafSource[curr_b_instance]; ok {
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

		// Jika ada meeting point ditemukan di level ini, kita bisa berhenti untuk mencari jalur terpendek.
		// Algoritma BFS dua arah standar biasanya berhenti di sini untuk menjamin jalur terpendek pertama.
		if len(meetingPoints) > 0 {
			break
		}
	}

	if len(meetingPoints) == 0 {
		return nil, -1, "", nil, exploredNodesCount
	}

	var bestMeetingPointData MeetingPoint
	foundBest := false
	bestTotalDepth := -1

	for _, mp := range meetingPoints {
		totalDepth := mp.forwardDepth + mp.backwardDepth
		if !foundBest || totalDepth < bestTotalDepth {
			bestTotalDepth = totalDepth
			bestMeetingPointData = mp
			foundBest = true
		} else if totalDepth == bestTotalDepth {
			// Tie-breaking: prefer path with shorter backward search (closer to the target leaf)
			// atau kriteria lain jika ada
			if bestMeetingPointData.actualTargetLeaf != nil && mp.actualTargetLeaf != nil {
				if mp.backwardDepth < bestMeetingPointData.backwardDepth {
					bestMeetingPointData = mp
				}
			}
		}
	}

	if !foundBest || bestMeetingPointData.actualTargetLeaf == nil { // Harus ada actualTargetLeaf yang valid
		return nil, -1, "", nil, exploredNodesCount
	}

	pathDesc := fmt.Sprintf("Path to %s (total depth: %d, fwd: %d, bwd: %d)",
		bestMeetingPointData.actualTargetLeaf.element, bestTotalDepth, bestMeetingPointData.forwardDepth, bestMeetingPointData.backwardDepth)

	// Menggunakan variabel global `recipeDatas`
	// Pastikan `recipeDatas.Recipes` adalah `map[string][][]string`
	// atau sesuaikan cara Anda mendapatkan `recipesMapConverted`
	var recipesMapConverted map[string][][]string
	if recipeDatas.Recipes != nil {
		recipesMapConverted = recipeDatas.Recipes
	} else {
		recipesMapConverted = make(map[string][][]string) // Fallback jika nil
	}

	resultTree := constructShortestPathTree(bestMeetingPointData.node, visited_f, visited_b, recipesMapConverted)

	if isPathCyclic(resultTree) {
		return nil, -1, "Cyclic path constructed", nil, exploredNodesCount
	}

	return resultTree, bestTotalDepth, pathDesc, bestMeetingPointData.actualTargetLeaf, exploredNodesCount
}

func searchBidirectionMultiple(target string, num int) ([]*Nodebidir, int) {
	fullTree := buildTreeBFS(target, recipeData.Recipes[target])

	pathTree, numPaths := findMultipleBidirectionalPaths(fullTree, num)

	if pathTree != nil {
		fmt.Println("\nPath Tree from BFS:")
		fmt.Printf("\nFound %d paths:\n", len(pathTree))
		for i, path := range pathTree {
			fmt.Printf("\nPath %d:\n", i+1)
			printShortestPathTree(path, "", true)
		}

	} else {
		fmt.Println("no valid path")
	}
	return pathTree, numPaths
}
