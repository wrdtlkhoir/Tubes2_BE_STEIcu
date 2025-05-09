package main

import (
	"fmt"
)

/*** SINGLE RECIPE (Shortest) ***/
func searchBFSOne(tree *Tree) ([]string, int) {
	fmt.Println("start bfs single")
	_, cntNodes, recipes := bfsOne([]*Node{tree.root}, 1, []string{})
	return recipes, cntNodes
}

// helper
func bfsOne(queue []*Node, cntNodes int, result []string) ([]*Node, int, []string) {

	// loop while queue tidak kosong
	for len(queue) > 0 {
		// dequeue first el
		var node *Node
		node, queue = dequeue(queue)

		// DEBUG
		fmt.Printf("curr el: %s\n", node.element)

		// append current node to result
		result = append(result, node.element)
		cntNodes++

		// jika node adalah daun, cont
		if isLeaf(node) {
			continue
		}

		for _, recipe := range node.combinations {
			// add to queue
			ing1 := recipe.ingredient1
			ing2 := recipe.ingredient2
			queue = enqueue(queue, ing1)
			queue = enqueue(queue, ing2)
		}
	}
	return nil, cntNodes, result
}

// func main() {
// 	recipeData := map[string][][]string{
// 		"Brick":    {{"Mud", "Fire"}, {"Clay", "Stone"}},
// 		"Mud":      {{"Water", "Earth"}},
// 		"Clay":     {{"Mud", "Sand"}},
// 		"Stone":    {{"Lava", "Air"}, {"Earth", "Pressure"}},
// 		"Sand":     {{"Earth", "Fire"}},
// 		"Lava":     {{"Earth", "Fire"}},
// 		"Pressure": {{"Air", "Air"}},
// 	}

// 	target := "Brick"

// 	tree := initTree(target, recipeData)
// 	printTree(tree)

// 	/* Try Single Recipe */
// 	result, num := searchBFSOne(tree)
// 	fmt.Println(result)
// 	fmt.Printf("Num of visitted nodes: %d\n", num)

// 	// // Get all elements needed for each recipe
// 	// recipeElements := findAllRecipes(target, recipeData)

// 	// // Print all elements as arrays
// 	// printRecipeElementsArrays(recipeElements)

// 	// // Print in the requested format
// 	// fmt.Println("\nFlattened recipe paths:")
// 	// for _, recipe := range recipeElements {
// 	// 	fmt.Println(flattenRecipePath(recipe))
// 	// }
// }
