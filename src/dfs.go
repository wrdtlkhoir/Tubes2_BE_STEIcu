package main

import (
	"fmt"
	"sync"
)

/*  TO DO:
1. change dummy to recipeData.Recipes[target]
2. handle element loops (if it end up to its own)
3. shortest if one dfs -- harusnya udah, cek lagi aja.
*/

var dummy = map[string][][]string {
	"Brick"		: {{"Mud", "Fire"}, {"Clay", "Stone"}},
	"Mud"		: {{"Water", "Earth"}},
	"Clay"		: {{"Mud", "Sand"}},
	"Stone"		: {{"Lava", "Air"}, {"Earth", "Pressure"}},
	"Sand" 		: {{"Earth", "Fire"}},
	"Lava" 		: {{"Earth", "Fire"}},
	"Pressure" 	: {{"Air", "Air"}},
}

/*** SINGLE RECIPE (Shortest) ***/

func searchDFSOne(target string) ([]string, int) {
	tree := initTree(target, dummy)
	printTree(tree)

	fmt.Println("start dfs single")

	return dfsOne(tree.root, 0, []string{})
}

// helper
func dfsOne(node *Node, cntNodes int, result []string) ([]string, int) {
	result = append(result, node.element)

	fmt.Printf("iterasi ke-%d elemen %s\n", cntNodes, node.element)

	if isLeaf(node) {
		return result, cntNodes
	}

	// only look for satu path aja
	result, cntNodes = dfsOne(node.combinations[0].ingredient1, cntNodes+1, result)
	result, cntNodes = dfsOne(node.combinations[0].ingredient2, cntNodes+1, result)
	return result, cntNodes
}



/*** MULTIPLE RECIPE ***/

type PathNode struct {
	element     string      // element at this node
	ingredients []*PathNode // ingredients to make this element (0 or 2)
}

var cache sync.Map

func searchDFSMultiple(target string, numRecipe int) ([][]string, []int) {
	tree := initTree(target, dummy)

	fmt.Println("start dfs multiple")

	cache = sync.Map{}
	pathNodes := dfsMultiple(tree.root)
	var paths [][]string
	var countNode []int
	cntNode := 0
	cntRecipe := 0

	// convert pathnode sama count visitted nodes
	for _, pathNode := range pathNodes {
		path := ConvertPathNode(pathNode)
		cntNode += len(path)
		paths = append(paths,path)
		countNode = append(countNode, cntNode)
		cntRecipe ++;
		if (cntRecipe == numRecipe) {break}
	}
	return paths, countNode
}

func dfsMultiple(node *Node) []*PathNode {
	// cek udah pernah visit elemen ini atau blom
	if cached, found := cache.Load(node.element); found {
		return cached.([]*PathNode)
	}

	// klo base el/leaf lgsg return
	if isBase(node.element) || isLeaf(node) {
		return []*PathNode{{element: node.element}}
	}
	
	// var for multithreading
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allPathNodes []*PathNode

	// each combination processed concurently
	for _, recipe := range node.combinations {
		wg.Add(1)
		go func (r Recipe) {
			defer wg.Done()

			if r.ingredient1 == nil || r.ingredient2 == nil {
				return
			}

			// get path nodes
			firstPathNodes := dfsMultiple(r.ingredient1)
			secondPathNodes := dfsMultiple(r.ingredient2)

			// combine path nodes
			for _, firstNode := range firstPathNodes {
				for _, secondNode := range secondPathNodes {
					newPathNode := &PathNode {
						element: node.element,
						ingredients: []*PathNode{firstNode, secondNode},
					}

					mu.Lock()
					allPathNodes = append(allPathNodes, newPathNode)
					mu.Unlock()
				}
			}
		}(recipe)
	}
	
	wg.Wait()
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

	// convert each ingred
	firstPath := ConvertPathNode(pathNode.ingredients[0])
	result = append(result, firstPath...)
	secondPath := ConvertPathNode(pathNode.ingredients[1])
	result = append(result, secondPath...)
	
	return result
}


func main() {
	LoadRecipes("initial_recipes.json")
	recipes, numNodes := searchDFSOne("Brick")
	fmt.Println(recipes)
	fmt.Printf("nodes visited: %d\n", numNodes)

	recipes2, numNodes2 := searchDFSMultiple("Brick", 2)
	for i, recipe := range recipes2 {
		fmt.Print(recipe)
		fmt.Printf(" - %d\n", numNodes2[i])
	}
}