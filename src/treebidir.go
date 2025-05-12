package main

import (
	"container/list"
	"fmt"
)

type Recipebidir struct {
	ingredient1 *Nodebidir
	ingredient2 *Nodebidir
}

type Nodebidir struct {
	element      string
	combinations []Recipebidir
	parent       *Nodebidir
	isCycleNode  bool
}

type Treebidir struct {
	root *Nodebidir
}

func isAncestor(node *Nodebidir, targetElement string) bool {
	curr := node.parent
	for curr != nil {
		if curr.element == targetElement {
			return true
		}
		curr = curr.parent
	}
	return false
}

func buildTreeBFS(target string, recipeData map[string][][]string) *Treebidir {
	root := &Nodebidir{element: target, parent: nil}
	queue := list.New()
	queue.PushBack(root)

	for queue.Len() > 0 {
		frontElement := queue.Front()
		currentNode := frontElement.Value.(*Nodebidir)
		queue.Remove(frontElement)
		
		fmt.Println(currentNode.element)

		if currentNode.isCycleNode {
			continue
		}

		if isBase(currentNode.element) {
			continue
		}

		recipes, exists := recipeData[currentNode.element]
		if !exists || len(recipes) == 0 {
			continue
		}

		for _, combination := range recipes {
			if len(combination) != 2 {
				continue
			}

			ing1Name := combination[0]
			ing2Name := combination[1]
			recipe := Recipebidir{}

			if isAncestor(currentNode, ing1Name) {
				recipe.ingredient1 = &Nodebidir{element: ing1Name, parent: currentNode, isCycleNode: true}
			} else if isBase(ing1Name) {
				recipe.ingredient1 = &Nodebidir{element: ing1Name, parent: currentNode}
			} else {
				ing1Node := &Nodebidir{element: ing1Name, parent: currentNode}
				recipe.ingredient1 = ing1Node
				queue.PushBack(ing1Node)
			}

			if isAncestor(currentNode, ing2Name) {
				recipe.ingredient2 = &Nodebidir{element: ing2Name, parent: currentNode, isCycleNode: true}
			} else if isBase(ing2Name) {
				recipe.ingredient2 = &Nodebidir{element: ing2Name, parent: currentNode}
			} else {
				ing2Node := &Nodebidir{element: ing2Name, parent: currentNode}
				recipe.ingredient2 = ing2Node
				queue.PushBack(ing2Node)
			}

			currentNode.combinations = append(currentNode.combinations, recipe)
		}
	}
	return &Treebidir{root: root}
}

func printTreeHelperBidir(node *Nodebidir, prefix string, isLast bool) {
	if node == nil {
		return
	}

	fmt.Print(prefix)
	if isLast {
		fmt.Print("└── ")
		prefix += "    "
	} else {
		fmt.Print("├── ")
		prefix += "│   "
	}

	fmt.Print(node.element)
	if node.isCycleNode {
		fmt.Print(" (Cycle)")
	}
	fmt.Println()

	if node.isCycleNode {
		return
	}

	for _, recipe := range node.combinations {
		printTreeHelperBidir(recipe.ingredient1, prefix, false)
		printTreeHelperBidir(recipe.ingredient2, prefix, true)
	}
}



func printTreeBidir(t *Treebidir) {
	if t == nil || t.root == nil {
		fmt.Println("Tree is empty")
		return
	}
	printTreeHelperBidir(t.root, "", true)
}

func buildTreeBFS2(target string, recipeData map[string][][]string) *Treebidir {
	root := &Nodebidir{element: target, parent: nil}
	queue := list.New()
	queue.PushBack(root)
	
	visited := make(map[string]bool)
	visited[target] = true
	
	processedCount := 0
	maxProcessed := 10000
	
	for queue.Len() > 0 && processedCount < maxProcessed {
		frontElement := queue.Front()
		currentNode := frontElement.Value.(*Nodebidir)
		queue.Remove(frontElement)
		
		processedCount++
		
		depth := getNodeDepth(currentNode)
		fmt.Printf("Processing: %s (depth: %d, processed: %d)\n", currentNode.element, depth, processedCount)
		
		if currentNode.isCycleNode {
			continue
		}
		
		if isBase(currentNode.element) {
			continue
		}
		
		recipes, exists := recipeData[currentNode.element]
		if !exists || len(recipes) == 0 {
			continue
		}
		
		for _, combination := range recipes {
			if len(combination) != 2 {
				continue
			}
			
			ing1Name := combination[0]
			ing2Name := combination[1]
			recipe := Recipebidir{}
			
			if isAncestorOptimized(currentNode, ing1Name) {
				recipe.ingredient1 = &Nodebidir{element: ing1Name, parent: currentNode, isCycleNode: true}
			} else if isBase(ing1Name) {
				recipe.ingredient1 = &Nodebidir{element: ing1Name, parent: currentNode}
			} else {
				ing1Node := &Nodebidir{element: ing1Name, parent: currentNode}
				recipe.ingredient1 = ing1Node
				
				if !visited[ing1Name] {
					visited[ing1Name] = true
					queue.PushBack(ing1Node)
				}
			}
			
			if isAncestorOptimized(currentNode, ing2Name) {
				recipe.ingredient2 = &Nodebidir{element: ing2Name, parent: currentNode, isCycleNode: true}
			} else if isBase(ing2Name) {
				recipe.ingredient2 = &Nodebidir{element: ing2Name, parent: currentNode}
			} else {
				ing2Node := &Nodebidir{element: ing2Name, parent: currentNode}
				recipe.ingredient2 = ing2Node
				
				if !visited[ing2Name] {
					visited[ing2Name] = true
					queue.PushBack(ing2Node)
				}
			}
			
			currentNode.combinations = append(currentNode.combinations, recipe)
		}
	}
	
	if processedCount >= maxProcessed {
		fmt.Println("tree building stopped. reaching maximum number of nodes")
	}
	
	return &Treebidir{root: root}
}

func getNodeDepth(node *Nodebidir) int {
	depth := 0
	current := node
	for current.parent != nil {
		depth++
		current = current.parent
	}
	return depth
}

func isAncestorOptimized(node *Nodebidir, elementName string) bool {
	current := node
	for current != nil {
		if current.element == elementName {
			return true
		}
		current = current.parent
	}
	return false
}