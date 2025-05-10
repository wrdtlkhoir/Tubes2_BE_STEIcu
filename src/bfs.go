package main

import (
	"fmt"
	"sync"
)

/*** SINGLE RECIPE BFS ***/

var memoSB map[string]*Node
var mainDataBFS map[string][][]string
var visitedBFS map[string]bool

func searchBFSOne(target string) (*Tree, int) {
	fmt.Println("start bfs single")
	memoSB = make(map[string]*Node)
	mainDataBFS = recipeData.Recipes[target]
	visitedBFS = make(map[string]bool)

	result, cntNode := bfsOne(target)

	return &Tree{root: result}, cntNode
}

func bfsOne(element string) (*Node, int) {
	cntNode := 0

	pendingNodes := make(map[string][]string)
	nodeMap := make(map[string]*Node)

	queue := []string{element}
	visitedBFS[element] = true

	for len(queue) > 0 {
		element := queue[0]
		queue = queue[1:]

		if _, exists := nodeMap[element]; !exists {
			nodeMap[element] = &Node{element: element}
		}

		if isBase(element) {
			continue
		}

		if ingredients, hasRecipe := mainDataBFS[element]; hasRecipe && len(ingredients) > 0 {
			pair := ingredients[0]
			pendingNodes[element] = pair

			for _, ingredient := range pair {
				cntNode++
				if !visitedBFS[ingredient] {
					visitedBFS[ingredient] = true
					queue = append(queue, ingredient)
				}
			}
		}
	}

	for el, ingredients := range pendingNodes {
		left := nodeMap[ingredients[0]]
		right := nodeMap[ingredients[1]]

		nodeMap[el].combinations = []Recipe{
			{
				ingredient1: left,
				ingredient2: right,
			},
		}

		memoSB[el] = nodeMap[el]
	}

	return nodeMap[element], cntNode
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