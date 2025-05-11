package main

import "fmt"

type Recipe struct {
	ingredient1 *Node
	ingredient2 *Node
}

type Node struct {
	element      string
	combinations []Recipe
}

type Tree struct {
	root *Node
}

// check if an element is base element
func isBase(element string) bool {
	return (element == "Air" || element == "Earth" || element == "Fire" || element == "Water")
}

// check if a node is a leaf
func isLeaf(node *Node) bool {
	return len(node.combinations) == 0
}

func isNodeUsedAsIngredient(target string, parent *Node) bool {
	if parent == nil {
		return false
	}
	for _, recipe := range parent.combinations {
		if recipe.ingredient1 != nil && recipe.ingredient1.element == target {
			return true
		}
		if recipe.ingredient2 != nil && recipe.ingredient2.element == target {
			return true
		}
		// recursively check children
		if isNodeUsedAsIngredient(target, recipe.ingredient1) || isNodeUsedAsIngredient(target, recipe.ingredient2) {
			return true
		}
	}
	return false
}

// build tree dari data recipe
var visited map[string]bool

func buildTree(target string, element string, recipeData map[string][][]string, cntNode int) *Node {
	if isBase(element) || (element == target && cntNode != 0) || visited[element] {
		return &Node{element: element}
	}

	visited[element] = true
	node := &Node{element: element}
	recipes := recipeData[element]

	for _, combination := range recipes {
		ing1 := buildTree(target, combination[0], recipeData, cntNode+1)
		ing2 := buildTree(target, combination[1], recipeData, cntNode+1)

		recipe := Recipe{
			ingredient1: ing1,
			ingredient2: ing2,
		}
		node.combinations = append(node.combinations, recipe)
	}
	return node
}

// called ini for init tree
func InitTree(target string, recipeData map[string][][]string) *Tree {
	visited = make(map[string]bool)
	root := buildTree(target, target, recipeData, 0)
	return &Tree{root: root}
}

func printTreeHelper(node *Node, prefix string, isLast bool) {
	if node == nil {
		return
	}

	// current node
	fmt.Print(prefix)
	if isLast {
		fmt.Print("└── ")
		prefix += "    "
	} else {
		fmt.Print("├── ")
		prefix += "│   "
	}
	fmt.Println(node.element)

	// combination
	for i, recipe := range node.combinations {

		printTreeHelper(recipe.ingredient1, prefix, false)

		isLastRecipe := i == len(node.combinations)-1
		printTreeHelper(recipe.ingredient2, prefix, isLastRecipe)
	}
}

// call this to print tree
func printTree(t *Tree) {
	printTreeHelper(t.root, "", true)
}
