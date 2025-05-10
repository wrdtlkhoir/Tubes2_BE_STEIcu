package main

import (
	"container/list"
	"fmt"
)

// Recipe represents a single combination of ingredients.
type Recipe struct {
	ingredient1 *Node
	ingredient2 *Node
}

// Node represents an element in the recipe tree.
type Node struct {
	element      string
	combinations []Recipe // Combinations needed to make this element

	parent *Node // Parent in the tree structure

	isCycleNode bool // Added: Flag to indicate this node represents a cycle point
}

// Tree represents the entire derivation tree starting from the root element.
type Tree struct {
	root *Node
}

// check if an element is base element
func isBase(element string) bool {
	baseElements := map[string]bool{
		"Air": true, "Earth": true, "Fire": true, "Water": true,
	}
	return baseElements[element]
}

// isAncestor checks if targetElement is an element of any ancestor node
// in the tree structure above the starting 'node'.
func isAncestor(node *Node, targetElement string) bool {
	// Start from the immediate parent
	curr := node.parent
	for curr != nil {
		if curr.element == targetElement {
			return true // Found the target element in the ancestry
		}
		// Move up to the next parent
		curr = curr.parent
	}
	return false // Target element not found in ancestry
}

// buildTreeBFS builds a *full* derivation tree using BFS.
// It detects cycles in branches and marks the node causing the cycle as a leaf.
func buildTreeBFS(target string, recipeData map[string][][]string) *Tree {
	// Create the root node for the target element
	root := &Node{element: target, parent: nil} // Root has no parent

	queue := list.New()
	queue.PushBack(root)

	fmt.Println("--- Starting BFS to build full tree for:", target, "---")
	fmt.Println("Initially enqueued:", root.element)
	fmt.Println("--- BFS Building Steps ---")

	// Perform BFS
	for queue.Len() > 0 {
		// Dequeue a node from the front of the queue
		frontElement := queue.Front()
		currentNode := frontElement.Value.(*Node)
		queue.Remove(frontElement)

		// Print the element being currently processed
		fmt.Println("Dequeued and processing:", currentNode.element)

		// If this node is marked as a cycle node, do not expand its recipes
		if currentNode.isCycleNode {
			fmt.Println("  ", currentNode.element, "is a cycle node, stopping expansion for this branch.")
			continue // Skip processing recipes for this node
		}

		// If the current node is a base element, it has no recipes to expand
		if isBase(currentNode.element) {
			fmt.Println("  ", currentNode.element, "is a base element, no further expansion needed.")
			continue
		}

		// Look up recipes for the current element
		recipes, exists := recipeData[currentNode.element]
		if !exists || len(recipes) == 0 {
			fmt.Println("  ", currentNode.element, "has no recipes defined (or recipes list is empty), stopping expansion for this branch.")
			continue
		}

		fmt.Println("  Found", len(recipes), "recipe(s) for", currentNode.element)

		// Process each possible recipe combination for the current element:
		for i, combination := range recipes {
			if len(combination) != 2 {
				fmt.Printf("    Warning: Skipping invalid recipe for %s (combination %d): %v (expected 2 ingredients)\n", currentNode.element, i+1, combination)
				continue
			}

			ing1Name := combination[0]
			ing2Name := combination[1]

			fmt.Printf("    Processing combination %d: [%s, %s]\n", i+1, ing1Name, ing2Name)

			// Create a Recipe instance for this combination
			recipe := Recipe{}
			// Track if we successfully created/enqueued at least one non-base, non-cycle ingredient node
			// in this combination to potentially add the recipe. (Decided earlier to add recipe always if combination exists)
			// Sticking to adding recipe always and handling nil/isCycleNode in print.

			// --- Process Ingredient 1 ---
			if isAncestor(currentNode, ing1Name) {
				// Cycle detected: ingredientName is an ancestor of currentNode
				fmt.Printf("      Cycle detected: '%s' is an ancestor of '%s'. Marking node as cycle leaf.\n", ing1Name, currentNode.element)
				// Create node, mark as cycle node, but DO NOT enqueue
				recipe.ingredient1 = &Node{element: ing1Name, parent: currentNode, isCycleNode: true}
			} else if isBase(ing1Name) {
				// It's a base element, create node but don't enqueue for further expansion
				recipe.ingredient1 = &Node{element: ing1Name, parent: currentNode}
				fmt.Println("      ", ing1Name, "is base, not enqueued.")
			} else {
				// Not base and not an ancestor, create node and enqueue for future expansion
				ing1Node := &Node{element: ing1Name, parent: currentNode}
				recipe.ingredient1 = ing1Node
				queue.PushBack(ing1Node)
				fmt.Println("      Enqueued:", ing1Node.element)
			}

			// --- Process Ingredient 2 ---
			if isAncestor(currentNode, ing2Name) {
				// Cycle detected
				fmt.Printf("      Cycle detected: '%s' is an ancestor of '%s'. Marking node as cycle leaf.\n", ing2Name, currentNode.element)
				// Create node, mark as cycle node, but DO NOT enqueue
				recipe.ingredient2 = &Node{element: ing2Name, parent: currentNode, isCycleNode: true}
			} else if isBase(ing2Name) {
				// It's a base element
				recipe.ingredient2 = &Node{element: ing2Name, parent: currentNode}
				fmt.Println("      ", ing2Name, "is base, not enqueued.")
			} else {
				// Not base and not an ancestor
				ing2Node := &Node{element: ing2Name, parent: currentNode}
				recipe.ingredient2 = ing2Node
				queue.PushBack(ing2Node)
				fmt.Println("      Enqueued:", ing2Node.element)
			}

			// Add the recipe combination to the current node's combinations
			// We always add the recipe struct if the combination exists in recipeData.
			currentNode.combinations = append(currentNode.combinations, recipe)

		}
		fmt.Println("  Finished processing recipes for:", currentNode.element)
	}

	fmt.Println("--- BFS building finished ---")
	return &Tree{root: root}
}

// --- Tree Printing Helper (Updated to handle cycle nodes) ---

func printTreeHelper(node *Node, prefix string, isLast bool) {
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

	// Print element, indicate if it's a cycle node
	fmt.Print(node.element)
	if node.isCycleNode {
		fmt.Print(" (Cycle)")
	}
	fmt.Println()

	// Do NOT print combinations if this is a cycle node
	if node.isCycleNode {
		return // Stop expansion down this branch
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
		// printTreeHelper handles nil ingredients by simply returning, so nothing is printed for nil branches.
		printTreeHelper(recipe.ingredient1, combinationChildPrefix, false) // Ingredient 1 is never the absolute last in its group combo print
		printTreeHelper(recipe.ingredient2, combinationChildPrefix, true)  // Ingredient 2 is always the last in its group combo print (for this recipe)
	}
}

// call this to print tree
func printTree(t *Tree) {
	if t == nil || t.root == nil {
		fmt.Println("Tree is empty")
		return
	}
	fmt.Println("\nRecipe Derivation Tree:")
	printTreeHelper(t.root, "", true)
}
