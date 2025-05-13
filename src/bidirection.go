package main

import (
	"container/list"
	"fmt"
)

func allLeavesAreBase(node *Nodebidir, visited map[*Nodebidir]bool) bool {
	if node == nil {
		return true
	}
	if visited[node] {
		return true
	}
	visited[node] = true

	if len(node.combinations) == 0 {
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

func findBaseLeaves(node *Nodebidir, baseLeaves []*Nodebidir) []*Nodebidir {
	if node == nil {
		return baseLeaves
	}

	if len(node.combinations) == 0 && isBase(node.element) {
		baseLeaves = append(baseLeaves, node)
	}

	for _, recipe := range node.combinations {
		if recipe.ingredient1 != nil && !recipe.ingredient1.isCycleNode {
			baseLeaves = findBaseLeaves(recipe.ingredient1, baseLeaves)
		}
		if recipe.ingredient2 != nil && !recipe.ingredient2.isCycleNode {
			baseLeaves = findBaseLeaves(recipe.ingredient2, baseLeaves)
		}
	}
	return baseLeaves
}

// VERSI DENGAN PEMROSESAN LEVEL LEBIH KETAT SEBELUM BREAK
func bidirectionalSearchTree(tree *Treebidir, recipesForItem map[string][][]string) (*Nodebidir, int) {
	exploredNodeCount := 0

	if tree == nil || tree.root == nil {
		return nil, exploredNodeCount
	}
	if tree.root.isCycleNode {
		return nil, exploredNodeCount
	}

	q_f := list.New()
	visited_f := make(map[*Nodebidir]*Nodebidir)
	root_f := tree.root
	q_f.PushBack(root_f)
	visited_f[root_f] = nil

	baseLeavesInitial := findBaseLeaves(tree.root, []*Nodebidir{})
	uniqueBaseLeavesMap := make(map[*Nodebidir]struct{})
	var baseLeaves []*Nodebidir
	for _, leaf := range baseLeavesInitial {
		if _, exists := uniqueBaseLeavesMap[leaf]; !exists {
			uniqueBaseLeavesMap[leaf] = struct{}{}
			baseLeaves = append(baseLeaves, leaf)
		}
	}

	q_b := list.New()
	visited_b := make(map[*Nodebidir]*Nodebidir)

	if len(baseLeaves) == 0 {
		if q_f.Len() == 1 && q_f.Front().Value.(*Nodebidir) == root_f {
			isRootBase := isBase(root_f.element)
			q_f.Remove(q_f.Front())
			exploredNodeCount = 1
			if isRootBase { // Jika root adalah base, itu adalah jalurnya
				return constructShortestPathTree(root_f, visited_f, visited_b, recipesForItem), exploredNodeCount
			}
		}
		return nil, exploredNodeCount // Tidak ada backward search atau hanya root non-base dieksplor
	}

	for _, baseLeaf := range baseLeaves {
		if baseLeaf == root_f {
			if q_f.Len() > 0 && q_f.Front().Value.(*Nodebidir) == root_f {
				q_f.Remove(q_f.Front())
				exploredNodeCount = 1
			} else if q_f.Len() == 0 { // Seharusnya tidak terjadi jika root_f ada
				exploredNodeCount = 0
			}
			return constructShortestPathTree(root_f, visited_f, visited_b, recipesForItem), exploredNodeCount
		}
		q_b.PushBack(baseLeaf)
		visited_b[baseLeaf] = nil
	}

	if q_b.Len() == 0 { // Tidak ada base leaves yang valid untuk memulai pencarian mundur setelah filter
		for q_f.Len() > 0 {
			q_f.Remove(q_f.Front())
			exploredNodeCount++
		}
		return nil, exploredNodeCount
	}

	forwardDepth := make(map[*Nodebidir]int)
	backwardDepth := make(map[*Nodebidir]int)
	forwardDepth[root_f] = 0
	for _, leaf := range baseLeaves {
		backwardDepth[leaf] = 0
	}

	type MeetingPointInfo struct {
		node          *Nodebidir
		forwardDepth  int
		backwardDepth int
	}
	var meetingPointsFound []MeetingPointInfo

	foundMeetingInThisPass := false

	for q_f.Len() > 0 && q_b.Len() > 0 {
		// Langkah Maju
		currentLevelSizeF := q_f.Len()
		for i := 0; i < currentLevelSizeF; i++ {
			frontElement_f := q_f.Front()
			curr_f_instance := frontElement_f.Value.(*Nodebidir)
			q_f.Remove(frontElement_f)
			exploredNodeCount++

			if curr_f_instance.isCycleNode {
				continue
			}

			if bDepth, found := backwardDepth[curr_f_instance]; found {
				meetingPointsFound = append(meetingPointsFound, MeetingPointInfo{
					curr_f_instance, forwardDepth[curr_f_instance], bDepth,
				})
				foundMeetingInThisPass = true // Tandai pertemuan
			}

			for _, recipe := range curr_f_instance.combinations {
				children := []*Nodebidir{recipe.ingredient1, recipe.ingredient2}
				for _, child_instance := range children {
					if child_instance != nil && !child_instance.isCycleNode {
						if _, visited := visited_f[child_instance]; !visited {
							q_f.PushBack(child_instance)
							visited_f[child_instance] = curr_f_instance
							newFwdDepth := forwardDepth[curr_f_instance] + 1
							forwardDepth[child_instance] = newFwdDepth
							if bDepthChild, found := backwardDepth[child_instance]; found {
								meetingPointsFound = append(meetingPointsFound, MeetingPointInfo{
									child_instance, newFwdDepth, bDepthChild,
								})
								foundMeetingInThisPass = true // Tandai pertemuan
							}
						}
					}
				}
			}
		}

		if foundMeetingInThisPass {
			break
		}

		currentLevelSizeB := q_b.Len()
		for i := 0; i < currentLevelSizeB; i++ {
			frontElement_b := q_b.Front()
			curr_b_instance := frontElement_b.Value.(*Nodebidir)
			q_b.Remove(frontElement_b)
			exploredNodeCount++

			if curr_b_instance.isCycleNode {
				continue
			}

			if fDepth, found := forwardDepth[curr_b_instance]; found {
				meetingPointsFound = append(meetingPointsFound, MeetingPointInfo{
					curr_b_instance, fDepth, backwardDepth[curr_b_instance],
				})
				foundMeetingInThisPass = true // Tandai pertemuan
			}

			parent_instance := curr_b_instance.parent
			if parent_instance != nil && !parent_instance.isCycleNode {
				if _, visited := visited_b[parent_instance]; !visited {
					q_b.PushBack(parent_instance)
					visited_b[parent_instance] = curr_b_instance
					newBwdDepth := backwardDepth[curr_b_instance] + 1
					backwardDepth[parent_instance] = newBwdDepth
					if fDepthParent, found := forwardDepth[parent_instance]; found {
						meetingPointsFound = append(meetingPointsFound, MeetingPointInfo{
							parent_instance, fDepthParent, newBwdDepth,
						})
						foundMeetingInThisPass = true // Tandai pertemuan
					}
				}
			}
		}

		if foundMeetingInThisPass {
			break
		}
	}

	if len(meetingPointsFound) == 0 {
		for q_f.Len() > 0 {
			q_f.Remove(q_f.Front())
			exploredNodeCount++
		}
		for q_b.Len() > 0 {
			q_b.Remove(q_b.Front())
			exploredNodeCount++
		}
		return nil, exploredNodeCount
	}

	var bestMeetingNode *Nodebidir
	minDepthSum := -1
	for _, mp := range meetingPointsFound {
		currentDepthSum := mp.forwardDepth + mp.backwardDepth
		if bestMeetingNode == nil || currentDepthSum < minDepthSum {
			minDepthSum = currentDepthSum
			bestMeetingNode = mp.node
		} else if currentDepthSum == minDepthSum {
			if mp.forwardDepth < forwardDepth[bestMeetingNode] {
				bestMeetingNode = mp.node
			}
		}
	}

	if bestMeetingNode != nil {
		pathTree := constructShortestPathTree(bestMeetingNode, visited_f, visited_b, recipesForItem)
		return pathTree, exploredNodeCount
	}
	return nil, exploredNodeCount
}

func constructShortestPathTree(meetingNode *Nodebidir, visited_f, visited_b map[*Nodebidir]*Nodebidir, recipeData map[string][][]string) *Nodebidir {
	forwardPath := []*Nodebidir{}
	curr := meetingNode
	for curr != nil {
		forwardPath = append([]*Nodebidir{curr}, forwardPath...)
		curr = visited_f[curr]
	}

	backwardPath := []*Nodebidir{}
	curr = meetingNode
	for curr != nil {
		backwardPath = append(backwardPath, curr)
		curr = visited_b[curr]
	}

	if len(backwardPath) > 0 {
		backwardPath = backwardPath[1:]
	}

	completePath := append(forwardPath, backwardPath...)
	return buildShortestPathTree(completePath, recipeData)
}

func buildShortestPathTree(path []*Nodebidir, recipeData map[string][][]string) *Nodebidir {
	if len(path) == 0 {
		return nil
	}

	nodeMap := make(map[*Nodebidir]*Nodebidir)

	for _, origNode := range path {
		nodeMap[origNode] = &Nodebidir{
			element:      origNode.element,
			combinations: []Recipebidir{},
		}
	}

	for i := 0; i < len(path)-1; i++ {
		origCurrent := path[i]
		origNext := path[i+1]

		nodeMap[origNext].parent = nodeMap[origCurrent]
	}

	expandNodeRecipes(path, nodeMap, recipeData)

	return nodeMap[path[0]]
}

// expandNodeRecipes mengisi kombinasi resep untuk setiap node dalam path yang sudah di-clone.
func expandNodeRecipes(path []*Nodebidir, nodeMap map[*Nodebidir]*Nodebidir, recipeData map[string][][]string) {
	isOriginalNodeActuallyInPath := func(nodeToTest *Nodebidir, currentLinearPath []*Nodebidir) bool {
		if nodeToTest == nil {
			return false
		}
		for _, pathNode := range currentLinearPath {
			if pathNode == nodeToTest {
				return true
			}
		}
		return false
	}

	for _, origNode := range path {
		if isBase(origNode.element) {
			continue
		}

		clonedNode := nodeMap[origNode]

		if len(origNode.combinations) > 0 {
			bestRecipe := findBestRecipe(origNode, path)

			if (bestRecipe.ingredient1 == nil || bestRecipe.ingredient1.isCycleNode) &&
				(bestRecipe.ingredient2 == nil || bestRecipe.ingredient2.isCycleNode) {
				continue
			}

			ingredient1Cloned := createOrGetIngredientNode(bestRecipe.ingredient1, nodeMap, clonedNode, path)
			ingredient2Cloned := createOrGetIngredientNode(bestRecipe.ingredient2, nodeMap, clonedNode, path)

			if ingredient1Cloned != nil || ingredient2Cloned != nil {
				clonedNode.combinations = append(clonedNode.combinations, Recipebidir{
					ingredient1: ingredient1Cloned,
					ingredient2: ingredient2Cloned,
				})
			}

			if ingredient1Cloned != nil && !isBase(ingredient1Cloned.element) &&
				(bestRecipe.ingredient1 != nil && !isOriginalNodeActuallyInPath(bestRecipe.ingredient1, path)) {
				expandIngredientRecursively(ingredient1Cloned, nodeMap, clonedNode, recipeData)
			}

			if ingredient2Cloned != nil && !isBase(ingredient2Cloned.element) &&
				(bestRecipe.ingredient2 != nil && !isOriginalNodeActuallyInPath(bestRecipe.ingredient2, path)) {
				expandIngredientRecursively(ingredient2Cloned, nodeMap, clonedNode, recipeData)
			}

		} else if origNode != path[len(path)-1] {
			recipes, exists := recipeData[origNode.element]
			if exists && len(recipes) > 0 {
				bestRecipeStrings := findBestRecipeFromData(origNode.element, recipes, path)
				if len(bestRecipeStrings) == 2 {
					ing1Node := &Nodebidir{
						element: bestRecipeStrings[0],
						parent:  clonedNode,
					}

					ing2Node := &Nodebidir{
						element: bestRecipeStrings[1],
						parent:  clonedNode,
					}

					clonedNode.combinations = append(clonedNode.combinations, Recipebidir{
						ingredient1: ing1Node,
						ingredient2: ing2Node,
					})
					if !isBase(ing1Node.element) {
						expandIngredientRecursively(ing1Node, nodeMap, clonedNode, recipeData)
					}
					if !isBase(ing2Node.element) {
						expandIngredientRecursively(ing2Node, nodeMap, clonedNode, recipeData)
					}
				}
			}
		}
	}
}

func findBestRecipeFromData(element string, recipes [][]string, path []*Nodebidir) []string {
	if len(recipes) == 0 {
		return nil
	}

	bestRecipe := recipes[0]
	bestScore := -1

	for _, recipe := range recipes {
		if len(recipe) != 2 {
			continue
		}

		score := 0

		if isBase(recipe[0]) {
			score += 5
		}

		if isBase(recipe[1]) {
			score += 5
		}

		if containsNodeByName(path, recipe[0]) {
			score += 10
		}

		if containsNodeByName(path, recipe[1]) {
			score += 10
		}

		if bestScore == -1 || score > bestScore {
			bestScore = score
			bestRecipe = recipe
		}
	}

	return bestRecipe
}

func findBestRecipe(node *Nodebidir, path []*Nodebidir) Recipebidir {
	if len(node.combinations) == 0 {
		return Recipebidir{}
	}

	var bestRecipe Recipebidir
	bestScore := -1

	for _, recipe := range node.combinations {
		if recipe.ingredient1 == nil || recipe.ingredient2 == nil {
			continue
		}

		if recipe.ingredient1.isCycleNode || recipe.ingredient2.isCycleNode {
			continue
		}

		score := 0

		if containsNode(path, recipe.ingredient1) {
			score += 10
		}

		if containsNode(path, recipe.ingredient2) {
			score += 10
		}

		if isBase(recipe.ingredient1.element) {
			score += 5
		}

		if isBase(recipe.ingredient2.element) {
			score += 5
		}
		if bestScore == -1 || score > bestScore {
			bestScore = score
			bestRecipe = recipe
		}
	}

	return bestRecipe
}

func containsNode(path []*Nodebidir, node *Nodebidir) bool {
	for _, pathNode := range path {
		if pathNode == node {
			return true
		}
	}
	return false
}

func containsNodeByName(path []*Nodebidir, element string) bool {
	for _, pathNode := range path {
		if pathNode.element == element {
			return true
		}
	}
	return false
}

func createOrGetIngredientNode(origIngredient *Nodebidir, nodeMap map[*Nodebidir]*Nodebidir, parent *Nodebidir, path []*Nodebidir) *Nodebidir {
	if origIngredient == nil {
		return nil
	}

	if origIngredient.isCycleNode {
		return nil
	}

	if cloned, exists := nodeMap[origIngredient]; exists {
		return cloned
	}

	clone := &Nodebidir{
		element:      origIngredient.element,
		parent:       parent,
		combinations: []Recipebidir{},
	}
	nodeMap[origIngredient] = clone

	return clone
}

func expandIngredientRecursively(node *Nodebidir, nodeMap map[*Nodebidir]*Nodebidir, parent *Nodebidir, recipeData map[string][][]string) {
	if node == nil || isBase(node.element) {
		return
	}

	var origNode *Nodebidir
	for origN, clonedN := range nodeMap {
		if clonedN == node {
			origNode = origN
			break
		}
	}

	if origNode != nil && len(origNode.combinations) > 0 {
		bestRecipe := findBestRecipe(origNode, []*Nodebidir{})

		if bestRecipe.ingredient1 == nil && bestRecipe.ingredient2 == nil {
			recipes, exists := recipeData[node.element]
			if exists && len(recipes) > 0 {
				bestRecipeData := findBestRecipeFromData(node.element, recipes, []*Nodebidir{})

				if len(bestRecipeData) == 2 {
					ing1Node := &Nodebidir{
						element: bestRecipeData[0],
						parent:  node,
					}

					ing2Node := &Nodebidir{
						element: bestRecipeData[1],
						parent:  node,
					}

					node.combinations = append(node.combinations, Recipebidir{
						ingredient1: ing1Node,
						ingredient2: ing2Node,
					})

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

		ingredient1 := createOrGetIngredientNode(bestRecipe.ingredient1, nodeMap, node, []*Nodebidir{})
		ingredient2 := createOrGetIngredientNode(bestRecipe.ingredient2, nodeMap, node, []*Nodebidir{})

		if ingredient1 != nil || ingredient2 != nil {
			node.combinations = append(node.combinations, Recipebidir{
				ingredient1: ingredient1,
				ingredient2: ingredient2,
			})

			if ingredient1 != nil && !isBase(ingredient1.element) {
				expandIngredientRecursively(ingredient1, nodeMap, node, recipeData)
			}

			if ingredient2 != nil && !isBase(ingredient2.element) {
				expandIngredientRecursively(ingredient2, nodeMap, node, recipeData)
			}
		}
	} else {
		recipes, exists := recipeData[node.element]
		if exists && len(recipes) > 0 {
			bestRecipe := findBestRecipeFromData(node.element, recipes, []*Nodebidir{})

			if len(bestRecipe) == 2 {
				ing1Node := &Nodebidir{
					element: bestRecipe[0],
					parent:  node,
				}

				ing2Node := &Nodebidir{
					element: bestRecipe[1],
					parent:  node,
				}

				node.combinations = append(node.combinations, Recipebidir{
					ingredient1: ing1Node,
					ingredient2: ing2Node,
				})

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

func printShortestPathTree(node *Nodebidir, prefix string, isLast bool) {
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

	fmt.Println(node.element)
	if isBase(node.element) {
		return
	}

	numCombinations := len(node.combinations)
	for i, recipe := range node.combinations {
		isLastCombination := (i == numCombinations-1)

		var combinationChildPrefix string
		fmt.Print(prefix)

		if isLastCombination {
			fmt.Print("└── ")
			combinationChildPrefix = prefix + "    "
		} else {
			fmt.Print("├── ")
			combinationChildPrefix = prefix + "│   "
		}

		fmt.Println()

		if recipe.ingredient1 != nil {
			printShortestPathTree(recipe.ingredient1, combinationChildPrefix, recipe.ingredient2 == nil)
		}

		if recipe.ingredient2 != nil {
			printShortestPathTree(recipe.ingredient2, combinationChildPrefix, true)
		}
	}
}

func searchBidirectOne(target string) (*Nodebidir, int) {
	recipesForTargetItem := recipeData.Recipes[target]
	fullTree := buildTreeBFS(target, recipesForTargetItem)
	pathTree, numPaths := bidirectionalSearchTree(fullTree, recipesForTargetItem)

	if pathTree != nil {
		printShortestPathTree(pathTree, "", true)
	}

	return pathTree, numPaths
}
