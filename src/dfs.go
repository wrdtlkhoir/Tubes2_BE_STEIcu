package main

import (
	"fmt"
	"sync"
)

/*** SINGLE RECIPE DFS ***/
var memoSD map[string]*Node
var mainData map[string][][]string
var visitedDFS map[string]bool
var currentPath map[string]bool

func searchDFSOne(target string) (*Tree, int) {
	fmt.Println("start dfs single")
	memoSD = make(map[string]*Node)
	mainData = recipeData.Recipes[target]
	visitedDFS = make(map[string]bool)
	currentPath = make(map[string]bool)
	result, cntNode, found := dfsOne(target, -1)

	fmt.Println("DFS done")
	for key := range currentPath {
		fmt.Printf("🚨 still in path: %s\n", key)
	}


	if found {
		return &Tree{root: result}, cntNode
	}
	return nil, 0
}

func dfsOne(element string, cntNode int) (*Node, int, bool) {
	// fmt.Println(element)
	cntNode++
	visitedDFS[element] = true
	if _, inPath := currentPath[element]; inPath {
		// fmt.Printf("el founded in current Path: %s\n", element)
		return nil, cntNode, false
	}

	currentPath[element] = true
	defer delete(currentPath, element)

	if isBase(element) {
		return &Node{element: element}, cntNode, true
	}

	if res, ok := memoSD[element]; ok {
		return res, cntNode, true
	}

	res := &Node{element: element, combinations: []Recipe{}}
	memoSD[element] = res

	if recipes, ok := mainData[element]; ok {
		temp := 0
		for _, ingredients := range recipes {
			temp++
			// fmt.Printf("element: %s, iterasi ke: %d\n", element, temp)

			left, updatedCntNode, leftValid := dfsOne(ingredients[0], cntNode)
			if !leftValid {
				continue
			}
			right, finalCntNode, rightValid := dfsOne(ingredients[1], updatedCntNode)
			if !rightValid {
				continue
			}
			res.combinations = append(res.combinations, Recipe{
				ingredient1: left,
				ingredient2: right,
			})
			cntNode = finalCntNode
			return res, cntNode, true
		}
	}
	return nil, cntNode, false
}

/*** MULTIPLE RECIPE DFS ***/
var mainDataMul map[string][][]string

func serializeTree(node *Node) string {
	if node == nil {
		return ""
	}
	if len(node.combinations) == 0 {
		return node.element
	}
	left := serializeTree(node.combinations[0].ingredient1)
	right := serializeTree(node.combinations[0].ingredient2)

	// Sort the left and right parts for canonical form
	if left > right {
		left, right = right, left
	}

	return fmt.Sprintf("%s(%s,%s)", node.element, left, right)
}

func searchDFSMultiple(target string, numOfPath int) ([]*Tree, []int) {
	fmt.Println("start dfs multiple")
	mainDataMul = recipeData.Recipes[target]
	rootNodes := dfsAll(target, numOfPath, mainDataMul)

	var trees []*Tree
	var pathElementCounts []int

	for _, rootNode := range rootNodes {
		trees = append(trees, &Tree{root: rootNode})
		pathElementCounts = append(pathElementCounts, getPathElementCount(rootNode))
	}
	return trees, pathElementCounts
}

func dfsAll(element string, numOfPath int, currentRecipeMap map[string][][]string) []*Node {
	fmt.Printf("dfs all: %s\n", element)
	var allFinalTargetTrees []*Node
	seenStructures := make(map[string]bool)

	targetCombs, exists := currentRecipeMap[element]
	if !exists || len(targetCombs) == 0 {
		return []*Node{{element: element}}
	}

	globalVisited := make(map[string]bool)
	globalVisited[element] = true

	for _, pair := range targetCombs {
		if len(pair) != 2 {
			continue
		}
		if pair[0] == element || pair[1] == element {
			fmt.Printf("Immediate loop detected with ingredient: %s\n", element)
			continue
		}

		var wg sync.WaitGroup
		var leftIngredientOptions, rightIngredientOptions []*Node
		var leftMutex, rightMutex sync.Mutex

		wg.Add(2)
		go func(ingName string) {
			defer wg.Done()
			visitedForPath := copyVisitedMap(globalVisited)
			leftResults := dfsSubTree(ingName, currentRecipeMap, visitedForPath)
			leftMutex.Lock()
			leftIngredientOptions = leftResults
			leftMutex.Unlock()
		}(pair[0])

		go func(ingName string) {
			defer wg.Done()
			visitedForPath := copyVisitedMap(globalVisited)
			rightResults := dfsSubTree(ingName, currentRecipeMap, visitedForPath)
			rightMutex.Lock()
			rightIngredientOptions = rightResults
			rightMutex.Unlock()
		}(pair[1])
		wg.Wait()

		if len(leftIngredientOptions) == 0 || len(rightIngredientOptions) == 0 {
			fmt.Printf("Skipping invalid path for %s: left=%d, right=%d\n", element, len(leftIngredientOptions), len(rightIngredientOptions))
			continue
		}

		for _, lNode := range leftIngredientOptions {
			for _, rNode := range rightIngredientOptions {
				finalTreeRoot := &Node{
					element: element,
					combinations: []Recipe{{
						ingredient1: lNode,
						ingredient2: rNode,
					}},
				}

				serialized := serializeTree(finalTreeRoot)
				if seenStructures[serialized] {
					continue
				}
				seenStructures[serialized] = true

				allFinalTargetTrees = append(allFinalTargetTrees, finalTreeRoot)
				if numOfPath > 0 && len(allFinalTargetTrees) >= numOfPath {
					return allFinalTargetTrees
				}
			}
		}
	}

	return allFinalTargetTrees
}


func dfsSubTree(element string, currentRecipeMap map[string][][]string, visitedFromCaller map[string]bool) []*Node {
	fmt.Printf("dfs subtree: %s\n", element)
	
	// Check if current element exists in the path so far
	if visitedFromCaller[element] {
		// fmt.Printf("Loop detected for element: %s\n", element)
		// This would create a loop, return empty to invalidate this path
		return []*Node{}
	}

	// Create a new visited map for the current path
	visitedForCurrentPath := make(map[string]bool)
	for k, v := range visitedFromCaller {
		visitedForCurrentPath[k] = v
	}
	visitedForCurrentPath[element] = true

	combs, exists := currentRecipeMap[element]
	if !exists || len(combs) == 0 {
		return []*Node{{element: element}}
	}

	var allPossibleLinearNodesForElement []*Node

	for _, pair := range combs {
		if len(pair) != 2 {
			continue
		}

		var wg sync.WaitGroup
		var leftIngredientOptions, rightIngredientOptions []*Node

		wg.Add(2)
		go func(ingName string) {
			defer wg.Done()
			leftIngredientOptions = dfsSubTree(ingName, currentRecipeMap, visitedForCurrentPath)
		}(pair[0])

		go func(ingName string) {
			defer wg.Done()
			rightIngredientOptions = dfsSubTree(ingName, currentRecipeMap, visitedForCurrentPath)
		}(pair[1])
		wg.Wait()

		// Only proceed if both paths are valid (no loops detected)
		if len(leftIngredientOptions) == 0 || len(rightIngredientOptions) == 0 {
			// Skip this combination if either path contains a loop
			fmt.Printf("Skipping invalid subpath for %s: left=%d, right=%d\n", 
				element, len(leftIngredientOptions), len(rightIngredientOptions))
			continue
		}

		for _, lNode := range leftIngredientOptions {
			for _, rNode := range rightIngredientOptions {
				newNode := &Node{
					element: element,
					combinations: []Recipe{{
						ingredient1: lNode,
						ingredient2: rNode,
					}},
				}
				allPossibleLinearNodesForElement = append(allPossibleLinearNodesForElement, newNode)
			}
		}
	}
	return allPossibleLinearNodesForElement
}

func countUniqueElementsInPath(node *Node, elements map[string]bool) {
	if node == nil {
		return
	}
	elements[node.element] = true
	if len(node.combinations) > 0 {
		recipe := node.combinations[0]
		countUniqueElementsInPath(recipe.ingredient1, elements)
		countUniqueElementsInPath(recipe.ingredient2, elements)
	}
}

func getPathElementCount(rootNode *Node) int {
	uniqueElements := make(map[string]bool)
	countUniqueElementsInPath(rootNode, uniqueElements)
	return len(uniqueElements)
}

func copyVisitedMap(original map[string]bool) map[string]bool {
	copied := make(map[string]bool, len(original))
	for k, v := range original {
		copied[k] = v
	}
	return copied
}


// func main() {
// 	loadRecipes("recipes.json")
// 	target := "Mud"
// 	numOfRecipe := 5

// 	// ini buat debug result aja
// 	// tree := InitTree(target, recipeData.Recipes[target])
// 	// printTree(tree)

// 	// Try Single Recipe
// 	result, nodes := searchDFSOne(target)
// 	printTree(result)
// 	if result.root == nil {
// 		fmt.Println("root is nil")
// 	}
// 	fmt.Printf("Number of visited nodes: %d\n", nodes)

// 	// Try multiple Recipe
// 	result2, nodes2 := searchDFSMultiple(target, numOfRecipe)
// 	for _, recipe := range result2 {
// 		printTree(recipe)
// 	}
// 	fmt.Printf("Number of visited nodes: %d\n", nodes2)
// }

