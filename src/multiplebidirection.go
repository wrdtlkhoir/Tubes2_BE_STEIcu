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

func findMultipleBidirectionalPaths(tree *Treebidir, numPaths int) []*Nodebidir {
	if tree == nil || tree.root == nil {
		fmt.Println("Tree is empty")
		return nil
	}

	if tree.root.isCycleNode {
		fmt.Println("Root is a cycle node")
		return nil
	}

	baseLeaves := findBaseLeaves(tree.root, []*Nodebidir{})
	if len(baseLeaves) == 0 {
		return nil
	}

	var validLeaves []*Nodebidir
	for _, leaf := range baseLeaves {
		if !leaf.isCycleNode {
			validLeaves = append(validLeaves, leaf)
		}
	}

	if len(validLeaves) == 0 {
		return nil
	}

	resultChan := make(chan PathResult, numPaths*2)
	var wg sync.WaitGroup
	var resultsMutex sync.Mutex
	var foundPaths []*Nodebidir
	var numFoundPaths int

	pathSignatures := make(map[string]bool)
	for i, singleBaseLeaf := range validLeaves {
		wg.Add(1)
		go func(leafIndex int, currentTargetLeaf *Nodebidir) {
			defer wg.Done()

			resultsMutex.Lock()
			shouldContinue := numFoundPaths < numPaths
			resultsMutex.Unlock()

			if !shouldContinue {
				return
			}
			if currentTargetLeaf.isCycleNode {
				return
			}

			path, score, desc, actualLeafFound := bidirectionalSearchFromLeaf(tree.root, []*Nodebidir{currentTargetLeaf})
			if path != nil && !isPathCyclic(path) {
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

		if isPathCyclic(result.path) {
			resultsMutex.Unlock()
			continue
		}
		if numFoundPaths < numPaths && !pathSignatures[result.pathSignature] {
			foundPaths = append(foundPaths, result.path)
			numFoundPaths++
			pathSignatures[result.pathSignature] = true
		}
		resultsMutex.Unlock()

		if numFoundPaths >= numPaths {
			break
		}
	}
	return foundPaths
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

func bidirectionalSearchFromLeaf(root *Nodebidir, targetBaseLeaves []*Nodebidir) (*Nodebidir, int, string, *Nodebidir) {
	type MeetingPoint struct {
		node             *Nodebidir
		forwardDepth     int
		backwardDepth    int
		actualTargetLeaf *Nodebidir
	}

	if root == nil {
		return nil, -1, "", nil
	}

	if len(targetBaseLeaves) == 0 {
		return nil, -1, "", nil
	}

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
	baseLeafSource := make(map[*Nodebidir]*Nodebidir)

	validTargetsFound := false
	for _, leaf := range targetBaseLeaves {
		if leaf == nil || leaf.isCycleNode {
			continue
		}
		q_b.PushBack(leaf)
		visited_b[leaf] = nil
		backwardDepth[leaf] = 0
		baseLeafSource[leaf] = leaf
		validTargetsFound = true
	}

	if !validTargetsFound {
		return nil, -1, "", nil
	}

	var meetingPoints []MeetingPoint

	for q_f.Len() > 0 && q_b.Len() > 0 {
		currLevelSize_f := q_f.Len()
		for i := 0; i < currLevelSize_f; i++ {
			frontElement_f := q_f.Front()
			curr_f_instance := frontElement_f.Value.(*Nodebidir)
			q_f.Remove(frontElement_f)
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
		currLevelSize_b := q_b.Len()
		for i := 0; i < currLevelSize_b; i++ {
			frontElement_b := q_b.Front()
			curr_b_instance := frontElement_b.Value.(*Nodebidir)
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
			bestMeetingPointData = mp
			foundBest = true
		} else if totalDepth == bestTotalDepth {
			if bestMeetingPointData.actualTargetLeaf != nil && mp.actualTargetLeaf != nil && // Ensure not nil
				mp.backwardDepth < bestMeetingPointData.backwardDepth {
				bestMeetingPointData = mp
			}
		}
	}

	if !foundBest || bestMeetingPointData.actualTargetLeaf == nil {
		return nil, -1, "", nil
	}

		pathDesc := fmt.Sprintf("Path to %s (total depth: %d)",
		bestMeetingPointData.actualTargetLeaf.element, bestTotalDepth)

	recipesMapConverted := map[string][][]string(recipeDatas.Recipes)
	resultTree := constructShortestPathTree(bestMeetingPointData.node, visited_f, visited_b, recipesMapConverted)
	if isPathCyclic(resultTree) {
		return nil, -1, "", nil
	}

	return resultTree, bestTotalDepth, pathDesc, bestMeetingPointData.actualTargetLeaf
}

func searchBidirectionMultiple(target string, num int) []*Nodebidir {
	fullTree := buildTreeBFS(target, recipeData.Recipes[target])

	pathTree := multipleBfsForBidir(fullTree, num)

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
	return pathTree
}
