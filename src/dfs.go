package main

import (
	"fmt"
	"sync"
	"time"
)

/*  TO DO:
1. change dummy to recipeData.Recipes[target]
2. handle element loops (if it end up to its own)
3. shortest if one dfs -- harusnya udah, cek lagi aja.
4. check number of visitted node
*/

var dummy = map[string][][]string {
	"Brick"		: {{"Mud", "Fire"}, {"Clay", "Stone"}},
	"Mud"		: {{"Water", "Earth"}},
	"Clay"		: {{"Mud", "Sand"}},
	"Stone"		: {{"Lava", "Air"}, {"Earth", "Pressure"}},
	"Sand" 		: {{"Stone", "Air"}},
	"Lava" 		: {{"Earth", "Fire"}},
	"Pressure" 	: {{"Air", "Air"}},
}

/*** SINGLE RECIPE (Shortest) ***/

func searchDFSOne(tree *Tree) ([]string, int) {
	fmt.Println("start dfs single")
	return dfsOne(tree.root, 1, []string{})
}

// helper
func dfsOne(node *Node, cntNodes int, result []string) ([]string, int) {
	// append current node to result
	result = append(result, node.element)

	// if Leaf, then searching for this path is stopped (returned)
	if isLeaf(node) {
		return result, cntNodes
	}

	// only look for satu path aja di masing-masing ingredient
	result, cntNodes = dfsOne(node.combinations[0].ingredient1, cntNodes+1, result)
	result, cntNodes = dfsOne(node.combinations[0].ingredient2, cntNodes+1, result)
	return result, cntNodes
}



/*** MULTIPLE RECIPE ***/

type PathNode struct {
	element     string      // element at this node
	ingredients []*PathNode // ingredients to make this element (0 or 2)
}

var cache sync.Map // concurrent map buat store info (cache) accross goroutines

func searchDFSMultiple(numRecipe int, tree *Tree) ([][]string, []int) {

	fmt.Println("start dfs multiple")

	cache = sync.Map{} // initialize dg Map kosong
	pathNodes := dfsMultiple(tree.root)

	var paths [][]string // storing path per recipe
	var countNode []int // storing visitted node per recipe
	cntNode := 0
	cntRecipe := 0

	// convert pathnode, count visitted nodes
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
	// cek apkh udah pernah visit elemen ini, klo udh langsung return pathnya
	if cached, found := cache.Load(node.element); found {
		return cached.([]*PathNode)
	}

	// klo base el/leaf lgsg return pathnode with only that element
	if isBase(node.element) || isLeaf(node) {
		return []*PathNode{{element: node.element}}
	}
	
	// var for multithreading
	var wg sync.WaitGroup // wait for all goroutines to finish
	var mu sync.Mutex // mutex to protect concurrent writes to allPathNodes
	var allPathNodes []*PathNode // buat store all recipe combinations

	// each combination processed concurently
	for _, recipe := range node.combinations {
		wg.Add(1) // add 1 ke waitgroup sblm launch goroutines
		// start goroutine for recipe 
		go func (r Recipe) {
			defer wg.Done() // ini buat ensure wg nya decrement pas udh selesai

			// skip klo nil
			if r.ingredient1 == nil || r.ingredient2 == nil {
				return
			}

			// get path nodes using dfs
			firstPathNodes := dfsMultiple(r.ingredient1)
			secondPathNodes := dfsMultiple(r.ingredient2)

			// combine path nodes
			for _, firstNode := range firstPathNodes {
				for _, secondNode := range secondPathNodes {
					newPathNode := &PathNode {
						element: node.element,
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
	
	wg.Wait() // waiting all goroutine to finish b4 continue
	cache.Store(node.element, allPathNodes) // store in cache
		
	return allPathNodes
}

func ConvertPathNode(pathNode *PathNode) []string {
	var result []string
	result = append(result, pathNode.element)
	
	// klo udh gaada ingredient, done
	if len(pathNode.ingredients) == 0 {
		return result
	}

	// convert each ingredient
	firstPath := ConvertPathNode(pathNode.ingredients[0])
	result = append(result, firstPath...)
	secondPath := ConvertPathNode(pathNode.ingredients[1])
	result = append(result, secondPath...)
	
	return result
}


func main() {
	// LoadRecipes("main-recipes.json")
	
	target := "Brick"

	tree := initTree(target, dummy)
	printTree(tree)

	/* Try Single Recipe */
	start := time.Now()
	recipes, numNodes := searchDFSOne(tree)
	duration := time.Since(start)

	fmt.Println(recipes)
	fmt.Printf("nodes visited: %d\n", numNodes)
	fmt.Printf("duration: %d ms\n", duration.Milliseconds())

	/* Try Multiple Recipe */
	startMul := time.Now()
	recipes2, numNodes2 := searchDFSMultiple(6, tree)
	durationMul := time.Since(startMul)

	for i, recipe := range recipes2 {
		fmt.Print(recipe)
		fmt.Printf(" - %d\n", numNodes2[i])
	}
	fmt.Printf("duration: %d ms\n", durationMul.Milliseconds())
}