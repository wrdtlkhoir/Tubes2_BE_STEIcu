package main

import (
	"container/list"
	"fmt"
	"strings"
)

type MeetingPoint struct {
	node          *Node
	forwardDepth  int
	backwardDepth int
	baseLeaf      *Node // Track which base leaf this path leads to
}

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

func bidirectionalSearchTree(tree *Tree) *Node {
	if tree == nil || tree.root == nil {
		fmt.Println("Tree is empty, cannot perform search.")
		return nil
	}

	// Skip if the root itself is a cycle node
	if tree.root.isCycleNode {
		fmt.Println("Root is a cycle node, cannot perform search.")
		return nil
	}

	// Forward search (starting from the root *instance* of the tree)
	q_f := list.New()
	// Visited map uses *Node instances* as keys for efficient lookup.
	visited_f := make(map[*Node]*Node) // node instance -> parent node instance in search path

	root_f := tree.root // Start search from the root Node *instance* of the tree
	q_f.PushBack(root_f)
	visited_f[root_f] = nil // The root of the search path has no parent in the search path

	// Backward search (starting from all base leaf *Node instances* in the tree)
	baseLeaves := findBaseLeaves(tree.root, []*Node{}) // Find the base leaf instances in the tree
	q_b := list.New()
	// Visited map uses *Node instances* as keys
	visited_b := make(map[*Node]*Node) // node instance -> parent node instance in search path

	if len(baseLeaves) == 0 {
		fmt.Println("No base leaves found in the tree, cannot perform backward search on tree.")
		return nil // Indicate no bidirectional path found
	}

	for _, baseLeaf := range baseLeaves {
		q_b.PushBack(baseLeaf)    // Start backward search from each base leaf instance
		visited_b[baseLeaf] = nil // The starting nodes of the backward search path have no parent in that search path
	}

	// Meeting point tracking

	var meetingPoints []MeetingPoint

	fmt.Println("\n--- Starting Bidirectional Search on Tree Structure ---")
	fmt.Println("Forward queue starts with root instance:", root_f.element)
	fmt.Printf("Backward queue starts with %d base leaf instances: ", len(baseLeaves))
	// Print base leaf elements for debugging
	for i, leaf := range baseLeaves {
		fmt.Print(leaf.element)
		if i < len(baseLeaves)-1 {
			fmt.Print(", ")
		}
	}
	fmt.Println()

	// Track depth in both directions
	forwardDepth := make(map[*Node]int)
	backwardDepth := make(map[*Node]int)

	// Track which base leaf a node came from in backward search
	baseLeafSource := make(map[*Node]*Node)

	forwardDepth[root_f] = 0
	for _, leaf := range baseLeaves {
		backwardDepth[leaf] = 0
		baseLeafSource[leaf] = leaf // Each base leaf is its own source
	}

	// Perform Bi-BFS
	for q_f.Len() > 0 && q_b.Len() > 0 {
		// Step forward (expand children via combinations)
		if q_f.Len() > 0 {
			frontElement_f := q_f.Front()
			curr_f_instance := frontElement_f.Value.(*Node) // Get a node instance from the tree
			q_f.Remove(frontElement_f)

			fmt.Printf("F: Processing node %s\n", curr_f_instance.element)

			// Skip cycle nodes in the forward direction
			if curr_f_instance.isCycleNode {
				fmt.Printf("F: Skipping cycle node %s\n", curr_f_instance.element)
				continue
			}

			// Check for collision (comparing *Node instances* - memory addresses)
			if backDepth, found := backwardDepth[curr_f_instance]; found {
				fmt.Printf("F: Meeting point found at %s (forward depth: %d, backward depth: %d)\n",
					curr_f_instance.element, forwardDepth[curr_f_instance], backDepth)
				meetingPoints = append(meetingPoints, MeetingPoint{
					node:          curr_f_instance,
					forwardDepth:  forwardDepth[curr_f_instance],
					backwardDepth: backDepth,
					baseLeaf:      baseLeafSource[curr_f_instance],
				})
				// Continue searching for potentially better paths
			}

			// Expand forward: Add ingredient *Node instances* from combinations
			for _, recipe := range curr_f_instance.combinations {
				children := []*Node{recipe.ingredient1, recipe.ingredient2}

				for _, child_instance := range children {
					if child_instance == nil || child_instance.isCycleNode {
						continue // Skip nil nodes and cycle nodes
					}

					// Skip if already visited in forward direction
					if _, v_found := visited_f[child_instance]; !v_found {
						fmt.Printf("F: Enqueueing child %s from %s\n", child_instance.element, curr_f_instance.element)
						q_f.PushBack(child_instance)
						visited_f[child_instance] = curr_f_instance // Record parent *instance* in search path
						forwardDepth[child_instance] = forwardDepth[curr_f_instance] + 1

						// Check for collision immediately
						if backDepth, b_found := backwardDepth[child_instance]; b_found {
							fmt.Printf("F: Immediate meeting point at %s (forward: %d, backward: %d)\n",
								child_instance.element, forwardDepth[child_instance], backDepth)
							meetingPoints = append(meetingPoints, MeetingPoint{
								node:          child_instance,
								forwardDepth:  forwardDepth[child_instance],
								backwardDepth: backDepth,
								baseLeaf:      baseLeafSource[child_instance],
							})
						}
					}
				}
			}
		}

		// Step backward (expand parent)
		if q_b.Len() > 0 {
			frontElement_b := q_b.Front()
			curr_b_instance := frontElement_b.Value.(*Node) // Get a node instance from the tree
			q_b.Remove(frontElement_b)

			fmt.Printf("B: Processing node %s\n", curr_b_instance.element)

			// Skip cycle nodes in the backward direction
			if curr_b_instance.isCycleNode {
				fmt.Printf("B: Skipping cycle node %s\n", curr_b_instance.element)
				continue
			}

			// Check for collision (comparing *Node instances*)
			if fwdDepth, found := forwardDepth[curr_b_instance]; found {
				fmt.Printf("B: Meeting point found at %s (forward: %d, backward: %d)\n",
					curr_b_instance.element, fwdDepth, backwardDepth[curr_b_instance])
				meetingPoints = append(meetingPoints, MeetingPoint{
					node:          curr_b_instance,
					forwardDepth:  fwdDepth,
					backwardDepth: backwardDepth[curr_b_instance],
					baseLeaf:      baseLeafSource[curr_b_instance],
				})
				// Continue searching for potentially better paths
			}

			// Expand backward: Move up to the parent *Node instance* in the tree structure
			parent_instance := curr_b_instance.parent // Use the parent link from the tree structure itself

			if parent_instance == nil || parent_instance.isCycleNode {
				continue // Skip nil parents and cycle nodes
			}

			// Check if this specific parent node instance has been visited in the backward search
			if _, v_found := visited_b[parent_instance]; !v_found {
				fmt.Printf("B: Enqueueing parent %s from %s\n", parent_instance.element, curr_b_instance.element)
				q_b.PushBack(parent_instance)
				visited_b[parent_instance] = curr_b_instance // Record child *instance* in search path
				backwardDepth[parent_instance] = backwardDepth[curr_b_instance] + 1

				// Propagate the base leaf source
				baseLeafSource[parent_instance] = baseLeafSource[curr_b_instance]

				// Check for collision immediately
				if fwdDepth, f_found := forwardDepth[parent_instance]; f_found {
					fmt.Printf("B: Immediate meeting point at %s (forward: %d, backward: %d)\n",
						parent_instance.element, fwdDepth, backwardDepth[parent_instance])
					meetingPoints = append(meetingPoints, MeetingPoint{
						node:          parent_instance,
						forwardDepth:  fwdDepth,
						backwardDepth: backwardDepth[parent_instance],
						baseLeaf:      baseLeafSource[parent_instance],
					})
				}
			}
		}
	}

	fmt.Println("--- Bidirectional Search on Tree Ended ---")
	fmt.Printf("Found %d meeting points\n", len(meetingPoints))

	// Find the meeting point with the shortest total path
	var bestMeetingPoint *MeetingPoint
	bestTotalDepth := -1

	for i := range meetingPoints {
		totalDepth := meetingPoints[i].forwardDepth + meetingPoints[i].backwardDepth
		fmt.Printf("Meeting point %d: %s (total depth: %d, leads to: %s)\n",
			i+1, meetingPoints[i].node.element, totalDepth,
			meetingPoints[i].baseLeaf.element)

		if bestTotalDepth == -1 || totalDepth < bestTotalDepth {
			bestTotalDepth = totalDepth
			bestMeetingPoint = &meetingPoints[i]
		}
	}

	if bestMeetingPoint == nil {
		fmt.Println("No valid meeting point found (all paths might contain cycle nodes)")
		return nil
	}

	fmt.Printf("Best meeting point: %s (forward depth: %d, backward depth: %d, total: %d)\n",
		bestMeetingPoint.node.element,
		bestMeetingPoint.forwardDepth,
		bestMeetingPoint.backwardDepth,
		bestTotalDepth)

	if bestMeetingPoint.baseLeaf != nil {
		fmt.Printf("This path leads to base element: %s\n", bestMeetingPoint.baseLeaf.element)
	}
	recipesMapConverted := map[string][][]string(recipeData.Recipes)
	// Create a subtree representing the shortest path
	return constructShortestPathTree(bestMeetingPoint.node, visited_f, visited_b, recipesMapConverted)
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
func mainWithShortestPathTree(recipeData map[string][][]string) {
	targetElement := "Swamp"
	// Build tree from data
	fullTree := buildTreeBFS(targetElement, recipeData)

	// Print the full tree
	fmt.Println("\nFull Recipe Derivation Tree:")
	printTree(fullTree)

	// Perform bidirectional search and get the shortest path tree
	shortestPathTree := bidirectionalSearchTree(fullTree)

	if shortestPathTree != nil {
		fmt.Println("\nShortest Path Tree:")
		printShortestPathTree(shortestPathTree, "", true)
	} else {
		fmt.Println("Could not find a valid path without cycles.")
	}

	fmt.Println("\n" + strings.Repeat("=", 40)) // Separator

}

// Function to replace main() for testing
func main() {
	recipeData := map[string][][]string{
		"Mud":      {{"Water", "Steam"}, {"Energy", "Water"}, {"Earth", "Earth"}},
		"Steam":    {{"Water", "Fire"}, {"Lava", "Fire"}},
		"Lava":     {{"Dust", "Fire"}, {"Water", "Fire"}},
		"Dust":     {{"Steam", "Air"}},
		"Energy":   {{"Air", "Fire"}},
		"Cloud":    {{"Air", "Water"}},
		"Swamp":    {{"Mud", "Water"}}, // Reordered to prioritize the correct/shorter path
		"Glass":    {{"Sand", "Fire"}, {"Fire", "Fire"}},
		"Sand":     {{"Stone", "Air"}},
		"Stone":    {{"Lava", "Water"}},
		"Obsidian": {{"Lava", "Water"}},
	}

	// mainWithShortestPathTree(recipeData)

	numPaths := 3
	mainWithMultiplePaths(recipeData, numPaths)
}
