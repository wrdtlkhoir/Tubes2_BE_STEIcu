package main

import (
	"container/list"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// (Ensure isBase function is defined elsewhere)
// var baseElements = map[string]bool{"Earth": true, "Air": true, "Fire": true, "Water": true}
// func isBase(element string) bool { return baseElements[element] }

func allLeavesAreBase(node *Node, visited map[*Node]bool) bool {
	if node == nil {
		return true // A nil ingredient in a recipe doesn't invalidate the path by this rule
	}
	if visited[node] {
		return true // Already validated this node or currently validating it up the recursion stack
	}
	visited[node] = true

	if len(node.combinations) == 0 { // This is a leaf in the *constructed path tree*
		return isBase(node.element)
	}

	for _, recipe := range node.combinations {
		if recipe.ingredient1 != nil && !allLeavesAreBase(recipe.ingredient1, visited) {
			return false
		}
		if recipe.ingredient2 != nil && !allLeavesAreBase(recipe.ingredient2, visited) {
			return false
		}
	}
	return true
}

// type MeetingPoint struct {
// 	node          *Node
// 	forwardDepth  int
// 	backwardDepth int
// 	baseLeaf      *Node // Track which base leaf this path leads to
// }

// Helper to find all base leaf nodes in the tree structure.
// This function traverses the *built tree* to find search starting points.
func findBaseLeaves(node *Node, baseLeaves []*Node) []*Node {
	if node == nil {
		return baseLeaves
	}

	// Check if it's a leaf node (no combinations) AND its element is a base element.
	if len(node.combinations) == 0 && isBase(node.element) {
		baseLeaves = append(baseLeaves, node)
	}

	// Recurse into combinations (ingredients Node instances)
	for _, recipe := range node.combinations {
		// Skip cycle nodes completely
		if recipe.ingredient1 != nil && !recipe.ingredient1.isCycleNode {
			baseLeaves = findBaseLeaves(recipe.ingredient1, baseLeaves)
		}
		if recipe.ingredient2 != nil && !recipe.ingredient2.isCycleNode {
			baseLeaves = findBaseLeaves(recipe.ingredient2, baseLeaves)
		}
	}
	return baseLeaves
}

func bidirectionalSearchTree(tree *Tree, recipesForItem map[string][][]string) *Node { // Added recipesForItem
	if tree == nil || tree.root == nil {
		fmt.Println("Tree is empty, cannot perform search.")
		return nil
	}
	if tree.root.isCycleNode {
		fmt.Println("Root is a cycle node, cannot perform search.")
		return nil
	}

	q_f := list.New()
	visited_f := make(map[*Node]*Node) // Store parent in path: child -> parent
	root_f := tree.root
	q_f.PushBack(root_f)
	visited_f[root_f] = nil

	baseLeaves := findBaseLeaves(tree.root, []*Node{})
	q_b := list.New()
	visited_b := make(map[*Node]*Node) // Store parent in path: child -> parent (for backward path reconstruction, this means it's child's child)

	if len(baseLeaves) == 0 {
		fmt.Println("No base leaves found, cannot perform backward search.")
		return nil
	}
	for _, baseLeaf := range baseLeaves {
		q_b.PushBack(baseLeaf)
		visited_b[baseLeaf] = nil
	}

	fmt.Println("\n--- Starting Bidirectional Search (seeking first constructible non-cyclic path) ---")
	// ... (your logging for queue starts) ...

	// Depth maps are still useful for understanding, though not for picking the "shortest" anymore
	forwardDepth := make(map[*Node]int)
	backwardDepth := make(map[*Node]int)
	baseLeafSource := make(map[*Node]*Node) // Still useful for context if needed

	forwardDepth[root_f] = 0
	for _, leaf := range baseLeaves {
		backwardDepth[leaf] = 0
		baseLeafSource[leaf] = leaf
	}

	// Perform Bi-BFS
	for q_f.Len() > 0 && q_b.Len() > 0 {
		// Step forward
		if q_f.Len() > 0 {
			frontElement_f := q_f.Front()
			curr_f_instance := frontElement_f.Value.(*Node)
			q_f.Remove(frontElement_f)

			fmt.Printf("F: Processing node %s (Depth: %d)\n", curr_f_instance.element, forwardDepth[curr_f_instance])
			if curr_f_instance.isCycleNode {
				fmt.Printf("F: Skipping cycle node %s\n", curr_f_instance.element)
				continue
			}

			// Check for collision
			if _, found := backwardDepth[curr_f_instance]; found {
				fmt.Printf("F: Meeting point candidate at %s. Attempting path construction...\n", curr_f_instance.element)
				// Attempt to construct the tree immediately using recipesForItem
				pathTree := constructShortestPathTree(curr_f_instance, visited_f, visited_b, recipesForItem)
				if pathTree != nil {
					fmt.Printf("F: Successfully constructed a path tree via meeting point %s. Returning.\n", curr_f_instance.element)
					return pathTree // Return the first successfully constructed path
				} else {
					fmt.Printf("F: Path construction failed for meeting point %s.\n", curr_f_instance.element)
				}
			}

			for _, recipe := range curr_f_instance.combinations {
				children := []*Node{recipe.ingredient1, recipe.ingredient2}
				for _, child_instance := range children {
					if child_instance == nil || child_instance.isCycleNode {
						continue
					}
					if _, v_found := visited_f[child_instance]; !v_found {
						fmt.Printf("F: Enqueueing child %s from %s\n", child_instance.element, curr_f_instance.element)
						q_f.PushBack(child_instance)
						visited_f[child_instance] = curr_f_instance
						forwardDepth[child_instance] = forwardDepth[curr_f_instance] + 1

						// Check for immediate collision after adding child
						if _, b_found := backwardDepth[child_instance]; b_found {
							fmt.Printf("F: Immediate meeting point candidate at enqueued child %s. Attempting path construction...\n", child_instance.element)
							pathTree := constructShortestPathTree(child_instance, visited_f, visited_b, recipesForItem)
							if pathTree != nil {
								fmt.Printf("F: Successfully constructed a path tree via immediate meeting at %s. Returning.\n", child_instance.element)
								return pathTree
							} else {
								fmt.Printf("F: Path construction failed for immediate meeting point %s.\n", child_instance.element)
							}
						}
					}
				}
			}
		}

		// Step backward
		if q_b.Len() > 0 {
			frontElement_b := q_b.Front()
			curr_b_instance := frontElement_b.Value.(*Node)
			q_b.Remove(frontElement_b)

			fmt.Printf("B: Processing node %s (Depth: %d)\n", curr_b_instance.element, backwardDepth[curr_b_instance])
			if curr_b_instance.isCycleNode {
				fmt.Printf("B: Skipping cycle node %s\n", curr_b_instance.element)
				continue
			}

			// Check for collision
			if _, found := forwardDepth[curr_b_instance]; found {
				fmt.Printf("B: Meeting point candidate at %s. Attempting path construction...\n", curr_b_instance.element)
				pathTree := constructShortestPathTree(curr_b_instance, visited_f, visited_b, recipesForItem)
				if pathTree != nil {
					fmt.Printf("B: Successfully constructed a path tree via meeting point %s. Returning.\n", curr_b_instance.element)
					return pathTree
				} else {
					fmt.Printf("B: Path construction failed for meeting point %s.\n", curr_b_instance.element)
				}
			}

			parent_instance := curr_b_instance.parent
			if parent_instance == nil || parent_instance.isCycleNode {
				continue
			}
			if _, v_found := visited_b[parent_instance]; !v_found {
				fmt.Printf("B: Enqueueing parent %s from %s\n", parent_instance.element, curr_b_instance.element)
				q_b.PushBack(parent_instance)
				visited_b[parent_instance] = curr_b_instance
				currentBaseLeaf := baseLeafSource[curr_b_instance] // Get base leaf from child
				baseLeafSource[parent_instance] = currentBaseLeaf  // Propagate to parent
				backwardDepth[parent_instance] = backwardDepth[curr_b_instance] + 1

				// Check for immediate collision after adding parent
				if _, f_found := forwardDepth[parent_instance]; f_found {
					fmt.Printf("B: Immediate meeting point candidate at enqueued parent %s. Attempting path construction...\n", parent_instance.element)
					pathTree := constructShortestPathTree(parent_instance, visited_f, visited_b, recipesForItem)
					if pathTree != nil {
						fmt.Printf("B: Successfully constructed a path tree via immediate meeting at %s. Returning.\n", parent_instance.element)
						return pathTree
					} else {
						fmt.Printf("B: Path construction failed for immediate meeting point %s.\n", parent_instance.element)
					}
				}
			}
		}
	}

	fmt.Println("--- Bidirectional Search Ended: No constructible non-cyclic path found ---")
	return nil // No path tree could be constructed from any meeting point
}

// Constructs a new tree representing the shortest path from root to base elements
func constructShortestPathTree(meetingNode *Node, visited_f, visited_b map[*Node]*Node, recipeData map[string][][]string) *Node {
	// First, trace path from meeting point to root (forward direction)
	forwardPath := []*Node{}
	curr := meetingNode
	for curr != nil {
		forwardPath = append([]*Node{curr}, forwardPath...) // Add to front
		curr = visited_f[curr]
	}

	// Trace path from meeting point to a base leaf (backward direction)
	backwardPath := []*Node{}
	curr = meetingNode
	for curr != nil {
		backwardPath = append(backwardPath, curr)
		curr = visited_b[curr]
	}

	// Remove duplicate meeting node from backward path
	if len(backwardPath) > 0 {
		backwardPath = backwardPath[1:]
	}

	// Combine paths: root -> meeting -> base
	completePath := append(forwardPath, backwardPath...)

	fmt.Println("\n--- Shortest Path Found ---")
	fmt.Print("Path: ")
	for i, node := range completePath {
		fmt.Print(node.element)
		if i < len(completePath)-1 {
			fmt.Print(" -> ")
		}
	}
	fmt.Println()

	// Now build a new tree representing just this path
	return buildShortestPathTree(completePath, recipeData)
}

// Builds a new tree with only the nodes in the shortest path
// Includes all recipes required to reach base elements
func buildShortestPathTree(path []*Node, recipeData map[string][][]string) *Node {
	if len(path) == 0 {
		return nil
	}

	// Create a mapping from original nodes to cloned nodes
	nodeMap := make(map[*Node]*Node)

	// First create clones for all nodes in the path
	for _, origNode := range path {
		nodeMap[origNode] = &Node{
			element:      origNode.element,
			combinations: []Recipe{},
		}
	}

	// Set up parent relationships for all nodes in the path
	for i := 0; i < len(path)-1; i++ {
		origCurrent := path[i]
		origNext := path[i+1]

		nodeMap[origNext].parent = nodeMap[origCurrent]
	}

	// Now recursively expand each non-base node in the path
	expandNodeRecipes(path, nodeMap, recipeData)

	return nodeMap[path[0]] // Return the root
}

// expandNodeRecipes expands the combinations for each node in the path
func expandNodeRecipes(path []*Node, nodeMap map[*Node]*Node, recipeData map[string][][]string) {
	for _, origNode := range path {
		// Skip base elements - they don't have recipes
		if isBase(origNode.element) {
			continue
		}

		clonedNode := nodeMap[origNode]

		// Find the best recipe for this node
		if len(origNode.combinations) > 0 {
			bestRecipe := findBestRecipe(origNode, path)

			// Skip if both ingredients are cycle nodes
			if (bestRecipe.ingredient1 == nil || bestRecipe.ingredient1.isCycleNode) &&
				(bestRecipe.ingredient2 == nil || bestRecipe.ingredient2.isCycleNode) {
				fmt.Printf("Warning: No valid recipe found for %s (all recipes contain cycle nodes)\n", origNode.element)
				continue
			}

			// Create recipe ingredients
			ingredient1 := createOrGetIngredientNode(bestRecipe.ingredient1, nodeMap, clonedNode, path)
			ingredient2 := createOrGetIngredientNode(bestRecipe.ingredient2, nodeMap, clonedNode, path)

			// Add recipe to the cloned node
			clonedNode.combinations = append(clonedNode.combinations, Recipe{
				ingredient1: ingredient1,
				ingredient2: ingredient2,
			})

			// Recursively expand the ingredients if they're not base elements and not already in the path
			if ingredient1 != nil && !isBase(ingredient1.element) && !containsNodeByName(path, ingredient1.element) {
				expandIngredientRecursively(ingredient1, nodeMap, clonedNode, recipeData)
			}

			if ingredient2 != nil && !isBase(ingredient2.element) && !containsNodeByName(path, ingredient2.element) {
				expandIngredientRecursively(ingredient2, nodeMap, clonedNode, recipeData)
			}
		} else if origNode != path[len(path)-1] { // Not the base leaf at the end
			// Try to look up a recipe from the recipe data if none found in the original node
			recipes, exists := recipeData[origNode.element]
			if exists && len(recipes) > 0 {
				// Find the best recipe from the data
				bestRecipe := findBestRecipeFromData(origNode.element, recipes, path)

				// Create the ingredient nodes
				ing1Node := &Node{
					element: bestRecipe[0],
					parent:  clonedNode,
				}

				ing2Node := &Node{
					element: bestRecipe[1],
					parent:  clonedNode,
				}

				// Add to the node map
				nodeMap[ing1Node] = ing1Node
				nodeMap[ing2Node] = ing2Node

				// Add the recipe to the cloned node
				clonedNode.combinations = append(clonedNode.combinations, Recipe{
					ingredient1: ing1Node,
					ingredient2: ing2Node,
				})
			}
		}
	}
}

// Finds the best recipe from recipe data for a node not in the original tree
func findBestRecipeFromData(element string, recipes [][]string, path []*Node) []string {
	if len(recipes) == 0 {
		return nil
	}

	// Start with the first recipe as the best
	bestRecipe := recipes[0]
	bestScore := -1

	for _, recipe := range recipes {
		if len(recipe) != 2 {
			continue // Skip invalid recipes
		}

		// Calculate score
		score := 0

		// Prefer base elements
		if isBase(recipe[0]) {
			score += 5
		}

		if isBase(recipe[1]) {
			score += 5
		}

		// Prefer elements already in the path
		if containsNodeByName(path, recipe[0]) {
			score += 10
		}

		if containsNodeByName(path, recipe[1]) {
			score += 10
		}

		// Update best recipe if this one has a higher score
		if bestScore == -1 || score > bestScore {
			bestScore = score
			bestRecipe = recipe
		}
	}

	return bestRecipe
}

// Finds the best recipe for a node, prioritizing recipes with ingredients in the path
// and completely avoiding recipes with cycle nodes
func findBestRecipe(node *Node, path []*Node) Recipe {
	if len(node.combinations) == 0 {
		return Recipe{} // Should not happen
	}

	// Start with the first recipe as the best
	var bestRecipe Recipe
	bestScore := -1

	for _, recipe := range node.combinations {
		// Skip recipes with nil ingredients
		if recipe.ingredient1 == nil || recipe.ingredient2 == nil {
			continue
		}

		// Skip recipes with cycle nodes completely
		if recipe.ingredient1.isCycleNode || recipe.ingredient2.isCycleNode {
			continue
		}

		// Calculate a score for this recipe
		// Higher score means it's preferred
		score := 0

		// Add points if ingredients are in the path
		if containsNode(path, recipe.ingredient1) {
			score += 10
		}

		if containsNode(path, recipe.ingredient2) {
			score += 10
		}

		// Add points if ingredients are base elements (prefer simpler recipes)
		if isBase(recipe.ingredient1.element) {
			score += 5
		}

		if isBase(recipe.ingredient2.element) {
			score += 5
		}

		// Update best recipe if this one has a higher score
		if bestScore == -1 || score > bestScore {
			bestScore = score
			bestRecipe = recipe
		}
	}

	return bestRecipe
}

// Check if a node is contained in the path
func containsNode(path []*Node, node *Node) bool {
	for _, pathNode := range path {
		if pathNode == node {
			return true
		}
	}
	return false
}

// Check if a node with the given element name is contained in the path
func containsNodeByName(path []*Node, element string) bool {
	for _, pathNode := range path {
		if pathNode.element == element {
			return true
		}
	}
	return false
}

// Returns an existing or creates a new ingredient node
// Will return nil for cycle nodes to avoid them in the result tree
func createOrGetIngredientNode(origIngredient *Node, nodeMap map[*Node]*Node, parent *Node, path []*Node) *Node {
	// Skip nil nodes
	if origIngredient == nil {
		return nil
	}

	// Skip cycle nodes completely - they should not be part of the result
	if origIngredient.isCycleNode {
		return nil
	}

	// If we already have a clone for this node, use it
	if cloned, exists := nodeMap[origIngredient]; exists {
		return cloned
	}

	// Otherwise create a new node
	clone := &Node{
		element:      origIngredient.element,
		parent:       parent,
		combinations: []Recipe{},
		// Never propagate cycle flags to the result tree
	}

	// Add to map
	nodeMap[origIngredient] = clone

	return clone
}

// Recursively expands an ingredient node and its children
func expandIngredientRecursively(node *Node, nodeMap map[*Node]*Node, parent *Node, recipeData map[string][][]string) {
	// Skip base elements or nil nodes
	if node == nil || isBase(node.element) {
		return
	}

	// Check map to find original node
	var origNode *Node
	for origN, clonedN := range nodeMap {
		if clonedN == node {
			origNode = origN
			break
		}
	}

	// If we found the original node and it has combinations
	if origNode != nil && len(origNode.combinations) > 0 {
		// Find the best recipe that doesn't include cycle nodes
		bestRecipe := findBestRecipe(origNode, []*Node{}) // Empty path since this is outside the main path

		// Skip if no valid recipe found (all recipes contain cycle nodes)
		if bestRecipe.ingredient1 == nil && bestRecipe.ingredient2 == nil {
			// Try to get a recipe from recipeData
			recipes, exists := recipeData[node.element]
			if exists && len(recipes) > 0 {
				// Get best recipe from data
				bestRecipeData := findBestRecipeFromData(node.element, recipes, []*Node{})

				if len(bestRecipeData) == 2 {
					// Create ingredient nodes
					ing1Node := &Node{
						element: bestRecipeData[0],
						parent:  node,
					}

					ing2Node := &Node{
						element: bestRecipeData[1],
						parent:  node,
					}

					// Add recipe to node
					node.combinations = append(node.combinations, Recipe{
						ingredient1: ing1Node,
						ingredient2: ing2Node,
					})

					// Recursively expand ingredients if needed
					if !isBase(ing1Node.element) {
						nodeMap[ing1Node] = ing1Node
						expandIngredientRecursively(ing1Node, nodeMap, node, recipeData)
					}

					if !isBase(ing2Node.element) {
						nodeMap[ing2Node] = ing2Node
						expandIngredientRecursively(ing2Node, nodeMap, node, recipeData)
					}
				}
			}
			return
		}

		// Create recipe ingredients
		ingredient1 := createOrGetIngredientNode(bestRecipe.ingredient1, nodeMap, node, []*Node{})
		ingredient2 := createOrGetIngredientNode(bestRecipe.ingredient2, nodeMap, node, []*Node{})

		// Add recipe to node if at least one ingredient is valid
		if ingredient1 != nil || ingredient2 != nil {
			node.combinations = append(node.combinations, Recipe{
				ingredient1: ingredient1,
				ingredient2: ingredient2,
			})

			// Recursively expand the ingredients
			if ingredient1 != nil && !isBase(ingredient1.element) {
				expandIngredientRecursively(ingredient1, nodeMap, node, recipeData)
			}

			if ingredient2 != nil && !isBase(ingredient2.element) {
				expandIngredientRecursively(ingredient2, nodeMap, node, recipeData)
			}
		}
	} else {
		// No original node found or it has no combinations
		// Try to get a recipe from recipeData
		recipes, exists := recipeData[node.element]
		if exists && len(recipes) > 0 {
			// Get best recipe from data
			bestRecipe := findBestRecipeFromData(node.element, recipes, []*Node{})

			if len(bestRecipe) == 2 {
				// Create ingredient nodes
				ing1Node := &Node{
					element: bestRecipe[0],
					parent:  node,
				}

				ing2Node := &Node{
					element: bestRecipe[1],
					parent:  node,
				}

				// Add recipe to node
				node.combinations = append(node.combinations, Recipe{
					ingredient1: ing1Node,
					ingredient2: ing2Node,
				})

				// Recursively expand ingredients if needed
				if !isBase(ing1Node.element) {
					nodeMap[ing1Node] = ing1Node
					expandIngredientRecursively(ing1Node, nodeMap, node, recipeData)
				}

				if !isBase(ing2Node.element) {
					nodeMap[ing2Node] = ing2Node
					expandIngredientRecursively(ing2Node, nodeMap, node, recipeData)
				}
			}
		}
	}
}

// printShortestPathTree prints the tree branch representing the shortest path
func printShortestPathTree(node *Node, prefix string, isLast bool) {
	if node == nil {
		return // Handles nil ingredient pointers
	}

	// Print the current node element
	fmt.Print(prefix)
	if isLast {
		fmt.Print("└── ")
		prefix += "    " // Extend prefix for children of the last item
	} else {
		fmt.Print("├── ")
		prefix += "│   " // Extend prefix for children of non-last item
	}

	// Print element
	fmt.Println(node.element)

	// If base element, stop here
	if isBase(node.element) {
		return
	}

	numCombinations := len(node.combinations)

	// Print combinations as groups branching off the parent
	for i, recipe := range node.combinations {
		isLastCombination := (i == numCombinations-1)

		// Calculate the prefix for the children nodes within this combination group.
		var combinationChildPrefix string
		fmt.Print(prefix) // Use the parent's child-line prefix

		if isLastCombination {
			fmt.Print("└── ")                        // Connector indicating this is the last combination group
			combinationChildPrefix = prefix + "    " // The vertical lines below this connector stop
		} else {
			fmt.Print("├── ")                        // Connector indicating this is not the last combination group
			combinationChildPrefix = prefix + "│   " // The vertical lines below this connector continue
		}
		// Print a newline after the connector to create the line segment
		fmt.Println()

		// Now print Ingredient 1 and Ingredient 2, indented under the combination group line.
		// Handle nil ingredients (which might happen if we filtered out cycle nodes)
		if recipe.ingredient1 != nil {
			printShortestPathTree(recipe.ingredient1, combinationChildPrefix, recipe.ingredient2 == nil)
		}

		if recipe.ingredient2 != nil {
			printShortestPathTree(recipe.ingredient2, combinationChildPrefix, true)
		}
	}
}

// Replacement for the main function to demonstrate the new implementation
// func mainWithShortestPathTree(recipeData map[string][][]string) {
// 	targetElement := "Swamp"
// 	// Build tree from data
// 	fullTree := buildTreeBFS(targetElement, recipeData)

// 	// Print the full tree
// 	fmt.Println("\nFull Recipe Derivation Tree:")
// 	printTree(fullTree)

// 	// Perform bidirectional search and get the shortest path tree
// 	shortestPathTree := bidirectionalSearchTree(fullTree)

// 	if shortestPathTree != nil {
// 		fmt.Println("\nShortest Path Tree:")
// 		printShortestPathTree(shortestPathTree, "", true)
// 	} else {
// 		fmt.Println("Could not find a valid path without cycles.")
// 	}

// 	fmt.Println("\n" + strings.Repeat("=", 40)) // Separator

// }

// Helper function to load OutputData from JSON
func LoadOutputDataFromJson(filename string) (OutputData, error) {
	var data OutputData
	jsonData, err := os.ReadFile(filename) // For Go 1.16+
	// For Go 1.15 and earlier, use: jsonData, err := ioutil.ReadFile(filename)
	if err != nil {
		return data, err
	}
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func main() {
	var allRecipeData OutputData

	log.Println("Loading recipe data from recipes.json...")
	loadedData, err := LoadOutputDataFromJson("recipes.json") // Ensure this file path is correct
	if err != nil {
		log.Fatalf("Error loading recipe data from JSON: %v", err)
		return
	}
	allRecipeData = loadedData
	log.Println("Recipe data loaded successfully.")

	if allRecipeData.Recipes == nil {
		log.Fatalln("Recipe data is empty after loading. Cannot proceed.")
		return
	}

	targetItemName := "Science" // Or your desired target

	// recipesForTargetItem will be of type map[string][][]string
	// This represents the categorized recipes for the targetItemName
	recipesForTargetItem, found := allRecipeData.Recipes[targetItemName]
	if !found {
		log.Printf("No recipes found for target item '%s' in allRecipeData.Recipes\n", targetItemName)
		available := []string{}
		for k := range allRecipeData.Recipes {
			available = append(available, k)
		}
		log.Printf("Available items in loaded data: %v", available)
		return
	}

	// Assuming buildTreeBFS uses recipesForTargetItem to build the initial, possibly pruned, tree.
	// buildTreeBFS needs to be robust in how it handles these recipes.
	fullTree := buildTreeBFS(targetItemName, recipesForTargetItem)
	if fullTree == nil || fullTree.root == nil {
		log.Fatalf("Failed to build the full tree for %s. It might be a base element or have no recipes.", targetItemName)
	}

	fmt.Println("\nFull Recipe Derivation Tree (output depends on your print function):")
	// print(fullTree) // Your placeholder for printing the full tree

	// Perform bidirectional search, passing the specific recipes for the target item
	fmt.Printf("\nStarting bidirectional search for %s (seeking first constructible path)...\n", targetItemName)
	// The recipesForTargetItem (map[string][][]string) is passed for path construction
	firstPathTree := bidirectionalSearchTree(fullTree, recipesForTargetItem)

	if firstPathTree != nil {
		fmt.Println("\n--- First Constructible Non-Cyclic Path Tree Found ---")
		printShortestPathTree(firstPathTree, "", true) // Using your existing print function for the result
	} else {
		fmt.Printf("\nCould not find any constructible non-cyclic path for '%s' via bidirectional search.\n", targetItemName)
	}

	fmt.Println("\n" + strings.Repeat("=", 40))
}
