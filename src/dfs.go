package main

import (
	"fmt"
	// "math"
	"sync"
	// "time"
)

/* TO DO */
// 1. Handle loop back

/*** SINGLE RECIPE DFS ***/

var memoSD map[string]*Node
var mainData map[string][][]string
var visitedDFS map[string]bool

func searchDFSOne(target string) (*Tree, int) {
	fmt.Println("start dfs single")
	memoSD = make(map[string]*Node)
	mainData = recipeData.Recipes[target]
	visitedDFS = make(map[string]bool)
	result, cntNode := dfsOne(target, -1)

	return &Tree{root: result}, cntNode
}

func dfsOne(element string, cntNode int) (*Node, int) {
	cntNode++
	visitedDFS[element] = true
	if isBase(element) {
		return &Node{element: element}, cntNode
	}
	if res, ok := memoSD[element]; ok {
		return res, cntNode
	}
	var res, left, right *Node
	if mainData[element] != nil {
		ingredients := mainData[element][0]
		// fmt.Printf("from: %s = %s - %s\n", element, ingredients[0], ingredients[1])
		left, cntNode = dfsOne(ingredients[0], cntNode)
		right, cntNode = dfsOne(ingredients[1], cntNode)

		res = &Node{
			element: element,
			combinations: []Recipe{
				{
					ingredient1: left,
					ingredient2: right,
				},
			},
		}
	}
	memoSD[element] = res
	return res, cntNode
}

/*** MULTIPLE RECIPE DFS ***/

var wg sync.WaitGroup
var mu sync.Mutex
var mainDataMul map[string][][]string

func searchDFSMultiple(target string, numOfPath int) ([]*Tree, []int) {
	fmt.Println("start dfs multiple")
	targetSpecificRecipes := recipeData.Recipes[target]
	rootNodes := dfsAll(target, numOfPath, targetSpecificRecipes)

	var trees []*Tree
	var pathElementCounts []int

	for _, rootNode := range rootNodes {
		trees = append(trees, &Tree{root: rootNode})
		pathElementCounts = append(pathElementCounts, getPathElementCount(rootNode))
	}
	return trees, pathElementCounts
}

func dfsAll(element string, numOfPath int, currentRecipeMap map[string][][]string) []*Node {
	var allFinalTargetTrees []*Node

	targetCombs, exists := currentRecipeMap[element]
	if !exists || len(targetCombs) == 0 {
		return []*Node{{element: element}}
	}

	for _, pair := range targetCombs {
		if len(pair) != 2 {
			continue
		}

		var wg sync.WaitGroup
		var leftIngredientOptions, rightIngredientOptions []*Node

		wg.Add(2)
		go func(ingName string) {
			defer wg.Done()
			visitedForPath := make(map[string]bool)
			visitedForPath[element] = true
			leftIngredientOptions = dfsSubTree(ingName, currentRecipeMap, visitedForPath)
		}(pair[0])

		go func(ingName string) {
			defer wg.Done()
			visitedForPath := make(map[string]bool)
			visitedForPath[element] = true
			rightIngredientOptions = dfsSubTree(ingName, currentRecipeMap, visitedForPath)
		}(pair[1])
		wg.Wait()

		if len(leftIngredientOptions) == 0 {
			leftIngredientOptions = []*Node{{element: pair[0]}}
		}
		if len(rightIngredientOptions) == 0 {
			rightIngredientOptions = []*Node{{element: pair[1]}}
		}

		for _, lNode := range leftIngredientOptions {
			for _, rNode := range rightIngredientOptions {
				if numOfPath > 0 && len(allFinalTargetTrees) >= numOfPath {
					goto endLoop // break out of all loops if enough paths are collected
				}
				finalTreeRoot := &Node{
					element: element,
					combinations: []Recipe{{
						ingredient1: lNode,
						ingredient2: rNode,
					}},
				}
				allFinalTargetTrees = append(allFinalTargetTrees, finalTreeRoot)
			}
		}
	}

endLoop:
	if numOfPath > 0 && len(allFinalTargetTrees) > numOfPath {
		return allFinalTargetTrees[:numOfPath]
	}
	return allFinalTargetTrees
}

func dfsSubTree(element string, currentRecipeMap map[string][][]string, visitedFromCaller map[string]bool) []*Node {
	if visitedFromCaller[element] {
		return []*Node{{element: element}}
	}

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

		if len(leftIngredientOptions) == 0 {
			leftIngredientOptions = []*Node{{element: pair[0]}}
		}
		if len(rightIngredientOptions) == 0 {
			rightIngredientOptions = []*Node{{element: pair[1]}}
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

// func main() {

// 	loadRecipes("filtered-recipe.json")
// 	target := "Clay"
// 	numOfRecipe := 3

// 	// ini buat debug result aja
// 	tree := InitTree(target, recipeData.Recipes[target])
// 	printTree(tree)

// 	// Try Single Recipe
// 	result, nodes := searchDFSOne(target)
// 	printTree(result)
// 	fmt.Printf("Number of visited nodes: %d\n", nodes)

// 	// Try multiple Recipe
// 	result2, nodes2 := searchDFSMultiple(target, numOfRecipe)
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