package main

import (
	"container/list"
	"fmt"
	"sync"
	"context"
	"time"
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

	pendingNodes := make(map[string][][]string)
	nodeMap := make(map[string]*Node)

	queue := []string{element}
	visitedBFS[element] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if _, exists := nodeMap[current]; !exists {
			nodeMap[current] = &Node{element: current}
		}

		if isBase(current) {
			continue
		}

		if recipes, hasRecipe := mainDataBFS[current]; hasRecipe && len(recipes) > 0 {
			for _, pair := range recipes {
				if len(pair) != 2 {
					continue
				}
				pendingNodes[current] = append(pendingNodes[current], pair)

				for _, ingredient := range pair {
					cntNode++
					if !visitedBFS[ingredient] {
						visitedBFS[ingredient] = true
						queue = append(queue, ingredient)
					}
				}
			}
		}
	}

	for el, allPairs := range pendingNodes {
		firstPair := allPairs[0]
		ing1 := nodeMap[firstPair[0]]
		ing2 := nodeMap[firstPair[1]]

		nodeMap[el].combinations = []Recipe{
			{
				ingredient1: ing1,
				ingredient2: ing2,
			},
		}

		memoSB[el] = nodeMap[el]
	}

	return nodeMap[element], cntNode
}


/*** MULTIPLE RECIPE BFS ***/
var allFinalTargetTrees []*Node

func searchBFSMultiple(target string, maxPathsToReturn int) ([]*Tree, []int) {
	fmt.Println("start bfs multiple")
	targetSpecificRecipes := recipeData.Recipes[target]
	rootNodes := bfsAll(target, maxPathsToReturn, targetSpecificRecipes)

	var trees []*Tree
	var pathElementCounts []int
	allFinalTargetTrees = make([]*Node, 0)

	for _, rootNode := range rootNodes {
		trees = append(trees, &Tree{root: rootNode})
		pathElementCounts = append(pathElementCounts, getPathElementCount(rootNode))
	}
	return trees, pathElementCounts
}


func bfsAll(targetElement string, maxPathsToReturn int, currentRecipeMap map[string][][]string) []*Node {
	var collectedTrees []*Node
	var mu sync.Mutex 

	targetTopLevelCombs, exists := currentRecipeMap[targetElement]
	if !exists || len(targetTopLevelCombs) == 0 {
		return []*Node{{element: targetElement}}
	}

	concurrencyLimit := 8 
	sem := make(chan struct{}, concurrencyLimit)
	resultChanBufferSize := 100
	if maxPathsToReturn > 0 && maxPathsToReturn < resultChanBufferSize {
		resultChanBufferSize = maxPathsToReturn
	}
	resultChan := make(chan *Node, resultChanBufferSize)
	
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pathsFound := 0

	for _, topRecipePair := range targetTopLevelCombs {
		if len(topRecipePair) != 2 {
			continue
		}

		sem <- struct{}{} 
		wg.Add(1)
		go func(recipe []string, currentContext context.Context) {
			defer func() {
				<-sem 
				wg.Done()
			}()

			// goroutineMemo := make(map[string][]*Node)
			pathVisited := make(map[string]bool)
			pathVisited[targetElement] = true

			ing1Name := recipe[0]
			ing2Name := recipe[1]

			expandedIng1Nodes := expandElementParallel(ing1Name, currentRecipeMap)
			expandedIng2Nodes := expandElementParallel(ing2Name, currentRecipeMap)
			
			delete(pathVisited, targetElement)

			if len(expandedIng1Nodes) == 0 { expandedIng1Nodes = []*Node{{element: ing1Name}} }
			if len(expandedIng2Nodes) == 0 { expandedIng2Nodes = []*Node{{element: ing2Name}} }

			for _, nodeIng1 := range expandedIng1Nodes {
				for _, nodeIng2 := range expandedIng2Nodes {
					select {
					case <-currentContext.Done():
						return
					default:
						// continue
					}

					rootNode := &Node{
						element: targetElement,
						combinations: []Recipe{{
							ingredient1: nodeIng1,
							ingredient2: nodeIng2,
						}},
					}

					select {
					case resultChan <- rootNode:
					case <-currentContext.Done():
						return
					}
				}
			}
		}(topRecipePair, ctx)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for tree := range resultChan {
		mu.Lock()
		if maxPathsToReturn <= 0 || pathsFound < maxPathsToReturn {
			collectedTrees = append(collectedTrees, tree)
			pathsFound++
			if maxPathsToReturn > 0 && pathsFound >= maxPathsToReturn {
				cancel() 
			}
		}
		mu.Unlock()
	}
	
	return collectedTrees
}

func expandElementParallel(
	elementName string,
	currentRecipeMap map[string][][]string,
) []*Node {
	var memo sync.Map
	var wg sync.WaitGroup
	jobChan := make(chan string, 100)

	worker := func() {
		for elem := range jobChan {
			// fmt.Println(elem)
			if _, exists := memo.Load(elem); exists {
				wg.Done()
				continue
			}

			recipes, ok := currentRecipeMap[elem]
			if !ok || len(recipes) == 0 {
				memo.Store(elem, []*Node{{element: elem}})
				wg.Done()
				continue
			}

			var nodes []*Node
			for _, recipe := range recipes {
				// fmt.Println("here")
				if len(recipe) != 2 {
					continue
				}
				ing1 := recipe[0]
				ing2 := recipe[1]

				// Add dependencies
				if _, ok := memo.Load(ing1); !ok {
					wg.Add(1)
					jobChan <- ing1
				}
				if _, ok := memo.Load(ing2); !ok {
					wg.Add(1)
					jobChan <- ing2
				}

				var ing1Nodes, ing2Nodes []*Node
				for {
					if val, ok := memo.Load(ing1); ok {
						ing1Nodes = val.([]*Node)
						break
					}
					time.Sleep(time.Millisecond)
				}
				for {
					if val, ok := memo.Load(ing2); ok {
						ing2Nodes = val.([]*Node)
						break
					}
					time.Sleep(time.Millisecond)
				}

				for _, n1 := range ing1Nodes {
					for _, n2 := range ing2Nodes {
						nodes = append(nodes, &Node{
							element: elem,
							combinations: []Recipe{{
								ingredient1: n1,
								ingredient2: n2,
							}},
						})
					}
				}
			}
			memo.Store(elem, nodes)
			wg.Done()
		}
	}

	// start workers
	for i := 0; i < 8; i++ {
		go worker()
	}

	wg.Add(1)
	jobChan <- elementName

	wg.Wait()
	close(jobChan)

	if val, ok := memo.Load(elementName); ok {
		return val.([]*Node)
	}
	return []*Node{}
}

func expandElement(
	elementName string,
	currentRecipeMap map[string][][]string,
	pathVisited map[string]bool,
	memo map[string][]*Node,
) []*Node {
	if nodes, found := memo[elementName]; found {
		return nodes
	}

	if pathVisited[elementName] {
		return []*Node{{element: elementName}}
	}
	pathVisited[elementName] = true
	defer delete(pathVisited, elementName)

	recipesForElement, exists := currentRecipeMap[elementName]
	if !exists || len(recipesForElement) == 0 {
		node := &Node{element: elementName}
		memo[elementName] = []*Node{node}
		return []*Node{node}
	}

	var allPossibleNodesForThisElement []*Node

	for _, recipePair := range recipesForElement {
		if len(recipePair) != 2 {
			continue
		}
		ing1Name := recipePair[0]
		ing2Name := recipePair[1]

		expandedIngredient1Nodes := expandElement(ing1Name, currentRecipeMap, pathVisited, memo)
		expandedIngredient2Nodes := expandElement(ing2Name, currentRecipeMap, pathVisited, memo)
		
		if len(expandedIngredient1Nodes) == 0 {
			expandedIngredient1Nodes = []*Node{{element: ing1Name}}
		}
		if len(expandedIngredient2Nodes) == 0 {
			expandedIngredient2Nodes = []*Node{{element: ing2Name}}
		}

		for _, nodeIng1 := range expandedIngredient1Nodes {
			for _, nodeIng2 := range expandedIngredient2Nodes {
				currentNode := &Node{
					element: elementName,
					combinations: []Recipe{{
						ingredient1: nodeIng1,
						ingredient2: nodeIng2,
					}},
				}
				allPossibleNodesForThisElement = append(allPossibleNodesForThisElement, currentNode)
			}
		}
	}

	memo[elementName] = allPossibleNodesForThisElement
	return allPossibleNodesForThisElement
}



/*** FOR BIDIRECTIONAL ***/
func multipleBfsForBidir(tree *Treebidir, numPaths int) []*Nodebidir {
	fmt.Println("start multiple bfs")

	if tree == nil || tree.root == nil {
		return nil
	}

	if tree.root.isCycleNode {
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

	for i, targetLeaf := range validLeaves {
		wg.Add(1)
		go func(leafIndex int, targetLeaf *Nodebidir) {
			defer wg.Done()
			resultsMutex.Lock()
			shouldContinue := numFoundPaths < numPaths
			resultsMutex.Unlock()

			if !shouldContinue {
				return
			}

			if targetLeaf.isCycleNode {
				return
			}

			path, score := bfsToTarget(tree.root, targetLeaf)

			if path != nil && !isPathCyclic(path) {
				pathSignature := generatePathSignature(path)

				resultChan <- PathResult{
					path:          path,
					score:         score,
					pathSignature: pathSignature,
				}
			} else if path != nil {
				// fmt.Printf("cyclic")
			}
		}(i, targetLeaf)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		resultsMutex.Lock()

		if isPathCyclic(result.path) {
			fmt.Printf("cyclic")
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

func bfsToTarget(root *Nodebidir, targetLeaf *Nodebidir) (*Nodebidir, int) {
	if root == nil || targetLeaf == nil {
		return nil, -1
	}

	if root.isCycleNode || targetLeaf.isCycleNode {
		return nil, -1
	}

	queue := list.New()
	visited := make(map[*Nodebidir]*Nodebidir)
	queue.PushBack(root)
	visited[root] = nil
	depth := make(map[*Nodebidir]int)
	depth[root] = 0

	foundTarget := false

	for queue.Len() > 0 && !foundTarget {
		current := queue.Front().Value.(*Nodebidir)
		queue.Remove(queue.Front())

		if current.isCycleNode {
			continue
		}

		if current == targetLeaf {
			foundTarget = true
			break
		}

		for _, recipe := range current.combinations {
			children := []*Nodebidir{recipe.ingredient1, recipe.ingredient2}
			for _, child := range children {
				if child == nil || child.isCycleNode {
					continue
				}

				if _, found := visited[child]; !found {
					queue.PushBack(child)
					visited[child] = current
					depth[child] = depth[current] + 1

					if child == targetLeaf {
						foundTarget = true
						break
					}
				}
			}
			if foundTarget {
				break
			}
		}
	}

	if !foundTarget {
		return nil, -1
	}
	resultTree := constructPathTree(targetLeaf, visited)

	return resultTree, depth[targetLeaf]
}

func constructPathTree(targetNode *Nodebidir, visited map[*Nodebidir]*Nodebidir) *Nodebidir {
	path := []*Nodebidir{}
	curr := targetNode
	for curr != nil {
		path = append([]*Nodebidir{curr}, path...)
		curr = visited[curr]
	}
	return buildShortestPathTree(path, recipeData.Recipes[targetNode.element])
}


// func main() {

// 	loadRecipes("recipes.json")
// 	target := "Rain"
// 	numOfRecipe := 2

// 	// ini buat debug result aja
// 	// tree := InitTree(target, recipeData.Recipes[target])
// 	// printTree(tree)

// 	// Try Single Recipe
// 	// result, nodes := searchBFSOne(target)
// 	// printTree(result)
// 	// fmt.Printf("Number of visited nodes: %d\n", nodes)

// 	// Try multiple Recipe
// 	result2, nodes2 := searchBFSMultiple(target, numOfRecipe)
// 	for _, recipe := range result2 {
// 		printTree(recipe)
// 	}
// 	fmt.Printf("Number of visited nodes: %d\n", nodes2)
// }