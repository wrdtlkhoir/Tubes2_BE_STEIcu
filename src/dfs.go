package main

import (
	"fmt"
	"math"
	"sync"
	// "time"
)

/*  TO DO:
2. handle element loops (if it end up to its own)
4. check number of visitted node
*/

var dummy = map[string][][]string{
	"Brick":    {{"Mud", "Fire"}, {"Clay", "Stone"}},
	"Mud":      {{"Water", "Earth"}},
	"Clay":     {{"Mud", "Sand"}},
	"Stone":    {{"Lava", "Air"}, {"Earth", "Pressure"}},
	"Sand":     {{"Stone", "Air"}},
	"Lava":     {{"Earth", "Fire"}},
	"Pressure": {{"Air", "Air"}},
}

/*** SINGLE RECIPE (Shortest) ***/

type Result struct {
	node      *Node
	stepCount int
}

var memoSD map[string]*Result
var numOfVisitedNodeSD int

func searchDFSOne(tree *Tree) *Node {
	fmt.Println("start dfs single")
	memoSD = make(map[string]*Result)
	numOfVisitedNodeSD = 0
	result := dfsOne(tree.root, 0)

	return result.node
}

func dfsOne(node *Node, cntNode int) *Result {

	if isBase(node.element) {
		return &Result{node: &Node{element: node.element}, stepCount: 0}
	}

	if res, ok := memoSD[node.element]; ok {
		return res
	}

	minSteps := math.MaxInt32
	var bestNode *Node

	for _, combo := range node.combinations {
		left := dfsOne(combo.ingredient1, cntNode+1)
		right := dfsOne(combo.ingredient2, cntNode+1)

		total := left.stepCount + right.stepCount + 1

		if total < minSteps {
			// build the current node with only this combination
			bestNode = &Node{
				element: node.element,
				combinations: []Recipe{
					{
						ingredient1: left.node,
						ingredient2: right.node,
					},
				},
			}
			minSteps = total
			numOfVisitedNodeSD = cntNode
		}
	}

	// save and return
	res := &Result{node: bestNode, stepCount: minSteps}
	memoSD[node.element] = res
	return res
}

/*** MULTIPLE RECIPE ***/

type PathNode struct {
	element     string      // element at this node
	ingredients []*PathNode // ingredients to make this element (0 or 2)
}

var cache sync.Map    // concurrent map to store cached results across goroutines
var visiting sync.Map // track currently processing elements to detect cycles

// searchDFSMultiple finds multiple recipe paths for the root element
func searchDFSMultiple(numRecipe int, tree *Tree) []*PathNode {
	fmt.Println("start dfs multiple")

	cache = sync.Map{}
	visiting = sync.Map{}

	// get all possible path nodes starting from root
	pathNodes := dfsMultiple(tree.root)

	return pathNodes
}

// dfsMultiple performs depth-first search to find all possible recipe paths
func dfsMultiple(node *Node) []*PathNode {
	if node == nil {
		return []*PathNode{}
	}

	if cached, found := cache.Load(node.element); found {
		return cached.([]*PathNode)
	}

	if isBase(node.element) {
		return []*PathNode{{element: node.element}}
	}

	visiting.Store(node.element, true)
	defer visiting.Delete(node.element)

	var wg sync.WaitGroup        // wait for all goroutines to finish
	var mu sync.Mutex            // mutex to protect concurrent writes
	var allPathNodes []*PathNode // store all recipe combinations

	for _, recipe := range node.combinations {
		wg.Add(1)
		go func(r Recipe) {
			defer wg.Done()

			if r.ingredient1 == nil || r.ingredient2 == nil {
				return
			}

			firstPathNodes := dfsMultiple(r.ingredient1)
			secondPathNodes := dfsMultiple(r.ingredient2)

			// combine paths from both ingredients
			for _, firstNode := range firstPathNodes {
				for _, secondNode := range secondPathNodes {
					newPathNode := &PathNode{
						element:     node.element,
					newPathNode := &PathNode{
						element:     node.element,
						ingredients: []*PathNode{firstNode, secondNode},
					}

					// safely append newPathNode
					mu.Lock()
					allPathNodes = append(allPathNodes, newPathNode)
					mu.Unlock()
				}
			}
		}(recipe)
	}

	wg.Wait()                               // waiting all goroutine to finish b4 continue
	cache.Store(node.element, allPathNodes) // store in cache


	wg.Wait() // wait for all goroutines to finish

	cache.Store(node.element, allPathNodes)
	return allPathNodes
}

func ConvertPathNode(pathNode *PathNode) []string {
	var result []string
	result = append(result, pathNode.element)

	// klo udh gaada ingredient, done
	if len(pathNode.ingredients) == 0 {
		return result
	}

	// convert each ingredient recipe, starting from ing0
	firstPath := ConvertPathNode(pathNode.ingredients[0])
	result = append(result, firstPath...)
	secondPath := ConvertPathNode(pathNode.ingredients[1])
	result = append(result, secondPath...)

func ConvertPathNodeToTree(pathNode *PathNode) map[string]interface{} {
	if pathNode == nil {
		return nil
	}

	result := map[string]interface{}{
		"element": pathNode.element,
	}

	if len(pathNode.ingredients) > 0 {
		ingredients := []map[string]interface{}{}

		for _, ing := range pathNode.ingredients {
			ingTree := ConvertPathNodeToTree(ing)
			if ingTree != nil {
				ingredients = append(ingredients, ingTree)
			}
		}

		if len(ingredients) > 0 {
			result["ingredients"] = ingredients
		}
	}

	return result
}


// func main() {
//     // Dummy target
//     target := "Brick"

//     // Inisialisasi tree
//     tree := initTree(target, dummy)

//     // Konversi tree ke JSON
//     treeJSON := convertToJSON(tree.root)

//     // Encode tree ke JSON dan cetak ke stdout
//     jsonData, err := json.MarshalIndent(treeJSON, "", "    ")
//     if err != nil {
//         log.Fatalf("Failed to encode tree to JSON: %v", err)
//     }

//     fmt.Println(string(jsonData))
// }

func main() {
	LoadFiltered("filtered-recipe.json")

	target := "Acid rain"
	mainRecipe = filteredData.Recipes[target]

	tree := initTreeDFS(target, mainRecipe)
	printTree(tree)

	result := searchDFSOne(tree) // ini return pointer to Node (which is tree hasil nya)
	printTreeHelper(result, "", true)
	fmt.Printf("Number of visited Node: %d\n", numOfVisitedNodeSD)

	// result2 := searchDFSMultiple(5, tree)

	// for _, path := range result2 {
	// 	PrintRecipeTree(path, " ")
	// }
}
