package main

import (
	"fmt"
	"time"
	"context"
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
	visitedNodes := make(map[string]bool)
	
	result, found := dfsOne(target, visitedNodes)

	visitedNodeCount := len(visitedNodes)
	fmt.Println("DFS done")
	if found {
		return &Tree{root: result}, visitedNodeCount
	}
	return nil, 0
}

func dfsOne(element string, visitedNodes map[string]bool) (*Node, bool) {
	visitedNodes[element] = true
	
	visitedDFS[element] = true
	if _, inPath := currentPath[element]; inPath {
		return nil, false
	}

	currentPath[element] = true
	defer delete(currentPath, element)

	if isBase(element) {
		return &Node{element: element}, true
	}

	if res, ok := memoSD[element]; ok {
		return res, true
	}

	res := &Node{element: element, combinations: []Recipe{}}
	memoSD[element] = res

	if recipes, ok := mainData[element]; ok {
		for _, ingredients := range recipes {
			left, leftValid := dfsOne(ingredients[0], visitedNodes)
			if !leftValid {
				continue
			}
			
			right, rightValid := dfsOne(ingredients[1], visitedNodes)
			if !rightValid {
				continue
			}
			
			res.combinations = append(res.combinations, Recipe{
				ingredient1: left,
				ingredient2: right,
			})
			
			return res, true
		}
	}
	return nil, false
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

	if left > right {
		left, right = right, left
	}

	return fmt.Sprintf("%s(%s,%s)", node.element, left, right)
}

func searchDFSMultiple(target string, numOfPath int) ([]*Tree, []int) {
	fmt.Println("start multiple")
	mainDataMul = recipeData.Recipes[target]
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	resultChan := make(chan *Node, numOfPath*2)
	var wg sync.WaitGroup
	
	var seenStructuresMutex sync.Mutex
	seenStructures := make(map[string]bool)
	
	targetCombs, exists := mainDataMul[target]
	if !exists || len(targetCombs) == 0 {
		return []*Tree{{root: &Node{element: target}}}, []int{1}
	}
	
	for _, pair := range targetCombs {
		if len(pair) != 2 {
			continue
		}
		if pair[0] == target || pair[1] == target {
			fmt.Printf("loop detected: %s\n", target)
			continue
		}
		
		wg.Add(1)
		go func(combo []string) {
			defer wg.Done()
			
			currentPath := make(map[string]bool)
			currentPath[target] = true
			
			leftPath := copyVisitedMap(currentPath)
			leftResults := dfsSubTree(ctx, combo[0], mainDataMul, leftPath, 0)
			
			if len(leftResults) == 0 {
				return
			}
			
			rightPath := copyVisitedMap(currentPath)
			rightResults := dfsSubTree(ctx, combo[1], mainDataMul, rightPath, 0)
			
			if len(rightResults) == 0 {
				return
			}
			
			maxCombos := 3
			for i := 0; i < min(len(leftResults), maxCombos); i++ {
				for j := 0; j < min(len(rightResults), maxCombos); j++ {
					select {
					case <-ctx.Done():
						return
					default:
						finalTreeRoot := &Node{
							element: target,
							combinations: []Recipe{{
								ingredient1: leftResults[i],
								ingredient2: rightResults[j],
							}},
						}
						
						serialized := serializeTree(finalTreeRoot)
						seenStructuresMutex.Lock()
						seen := seenStructures[serialized]
						if !seen {
							seenStructures[serialized] = true
							seenStructuresMutex.Unlock()
							
							select {
							case resultChan <- finalTreeRoot:
								// success
							default:
								// channel buffer full, continue
							}
						} else {
							seenStructuresMutex.Unlock()
						}
					}
				}
			}
		}(pair)
	}
	
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	var allResults []*Node
	for result := range resultChan {
		allResults = append(allResults, result)
		if len(allResults) >= numOfPath {
			break
		}
	}
	
	var trees []*Tree
	var pathElementCounts []int

	for _, rootNode := range allResults {
		trees = append(trees, &Tree{root: rootNode})
		pathElementCounts = append(pathElementCounts, getPathElementCount(rootNode))
	}
	
	fmt.Printf("found %d unique recipe paths\n", len(trees))
	return trees, pathElementCounts
}

func dfsSubTree(ctx context.Context, element string, currentRecipeMap map[string][][]string, currentPath map[string]bool, depth int) []*Node {
	// fmt.Println(element)
	select {
	case <-ctx.Done():
		return []*Node{} 
	default:
	}

	maxDepth := 15
	if depth > maxDepth {
		return []*Node{{element: element}}
	}
	
	if currentPath[element] {
		return []*Node{}
	}

	combs, exists := currentRecipeMap[element]
	if !exists || len(combs) == 0 || isBase(element) {
		return []*Node{{element: element}}
	}

	currentPath[element] = true
	
	if depth < 3 && len(combs) > 1 {
		var wg sync.WaitGroup
		resultsMutex := sync.Mutex{}
		var allPossibleNodesForElement []*Node
		
		maxCombsToExplore := min(len(combs), 2)
		combsToExplore := combs[:maxCombsToExplore]
		
		for _, pair := range combsToExplore {
			if len(pair) != 2 {
				continue
			}
			
			wg.Add(1)
			go func(ingredients []string) {
				defer wg.Done()
				
				select {
				case <-ctx.Done():
					return
				default:
				}
				
				leftPath := copyVisitedMap(currentPath)
				leftResults := dfsSubTree(ctx, ingredients[0], currentRecipeMap, leftPath, depth+1)
				
				if len(leftResults) == 0 {
					return
				}
				
				rightPath := copyVisitedMap(currentPath)
				rightResults := dfsSubTree(ctx, ingredients[1], currentRecipeMap, rightPath, depth+1)
				
				if len(rightResults) == 0 {
					return
				}
				
				var localNodes []*Node
				for i := 0; i < min(len(leftResults), 2); i++ {
					for j := 0; j < min(len(rightResults), 2); j++ {
						newNode := &Node{
							element: element,
							combinations: []Recipe{{
								ingredient1: leftResults[i],
								ingredient2: rightResults[j],
							}},
						}
						localNodes = append(localNodes, newNode)
						
						if len(localNodes) >= 3 {
							break
						}
					}
					if len(localNodes) >= 3 {
						break
					}
				}
				
				if len(localNodes) > 0 {
					resultsMutex.Lock()
					allPossibleNodesForElement = append(allPossibleNodesForElement, localNodes...)
					resultsMutex.Unlock()
				}
			}(pair)
		}
		
		wg.Wait()
		
		if len(allPossibleNodesForElement) > 5 {
			return allPossibleNodesForElement[:5]
		}
		return allPossibleNodesForElement
	} else {
		var allPossibleNodesForElement []*Node
		
		maxCombsToExplore := 2
		if len(combs) > maxCombsToExplore {
			combs = combs[:maxCombsToExplore]
		}
		
		for _, pair := range combs {
			if len(pair) != 2 {
				continue
			}

			leftPath := copyVisitedMap(currentPath)
			leftIngredientOptions := dfsSubTree(ctx, pair[0], currentRecipeMap, leftPath, depth+1)
			if len(leftIngredientOptions) == 0 {
				continue
			}
			
			rightPath := copyVisitedMap(currentPath)
			rightIngredientOptions := dfsSubTree(ctx, pair[1], currentRecipeMap, rightPath, depth+1)
			if len(rightIngredientOptions) == 0 {
				continue
			}

			maxOptions := 10
			leftLimit := min(len(leftIngredientOptions), maxOptions)
			rightLimit := min(len(rightIngredientOptions), maxOptions)
			
			for i := 0; i < leftLimit; i++ {
				for j := 0; j < rightLimit; j++ {
					newNode := &Node{
						element: element,
						combinations: []Recipe{{
							ingredient1: leftIngredientOptions[i],
							ingredient2: rightIngredientOptions[j],
						}},
					}
					allPossibleNodesForElement = append(allPossibleNodesForElement, newNode)
					
					if len(allPossibleNodesForElement) >= 5 {
						return allPossibleNodesForElement
					}
				}
			}
		}
		
		return allPossibleNodesForElement
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
// 	target := "Librarian"
// 	numOfRecipe := 2

// 	// ini buat debug result aja
// 	tree := InitTree(target, recipeData.Recipes[target])
// 	printTree(tree)

// 	// Try Single Recipe
// 	// result, nodes := searchDFSOne(target)
// 	// printTree(result)
// 	// if result.root == nil {
// 	// 	fmt.Println("root is nil")
// 	// }
// 	// fmt.Printf("Number of visited nodes: %d\n", nodes)

// 	// Try multiple Recipe
// 	result2, nodes2 := searchDFSMultiple(target, numOfRecipe)
// 	for _, recipe := range result2 {
// 		printTree(recipe)
// 	}
// 	fmt.Printf("Number of visited nodes: %d\n", nodes2)
// }