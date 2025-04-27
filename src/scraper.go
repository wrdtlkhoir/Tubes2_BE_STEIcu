package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const url = "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"

// Structured output format
type OutputData struct {
	Elements []string                         `json:"elements"` // List of all element names
	Recipes  map[string]map[string][][]string `json:"recipes"`  // The recipe data
}

// ScrapeRecursiveRecipes scrapes recipes recursively from the Little Alchemy 2 wiki.
func ScrapeInitialRecipes() (OutputData, error) {
	// Initialize the result structure with starting elements
	result := OutputData{
		Elements: []string{"Air", "Earth", "Fire", "Water"},
		Recipes:  make(map[string]map[string][][]string),
	}

	res, err := http.Get(url)
	if err != nil {
		return OutputData{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return OutputData{}, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return OutputData{}, err
	}

	// Step 1: Count how many tables match our criteria
	targetTables := []*goquery.Selection{}
	targetParagraphs := []*goquery.Selection{}

	doc.Find("p").Each(func(_ int, p *goquery.Selection) {
		if strings.Contains(p.Text(), "These elements can be created by combining only") {
			nextTable := p.NextFiltered("table")
			if nextTable.Length() > 0 {
				targetParagraphs = append(targetParagraphs, p)
				targetTables = append(targetTables, nextTable)
			}
		}
	})

	tableCount := len(targetTables)
	fmt.Printf("Found %d tables matching the criteria\n", tableCount)

	// // Limit to first 5 tables
	// maxTables := 10
	// if tableCount > maxTables {
	// 	tableCount = maxTables
	// 	targetTables = targetTables[:maxTables]
	// }

	// Step 2: Process tables one by one, building up available elements
	allElements := map[string]bool{
		"Air":   true,
		"Earth": true,
		"Fire":  true,
		"Water": true,
	}

	allRecipesUpToTable := make(map[int]map[string][][]string)
	elementsPerTable := make(map[int][]string)

	// Initially, allRecipesUpToTable[0] is empty (only base elements)
	allRecipesUpToTable[0] = make(map[string][][]string)

	// Process each table in sequence
	for i := 0; i < tableCount; i++ {
		currentTable := targetTables[i]

		// Store elements found in this table
		elementsInThisTable := []string{}
		recipesInThisTable := make(map[string][][]string)

		// Extract data from this table
		rows := currentTable.Find("tr").Slice(1, currentTable.Find("tr").Length())
		rows.Each(func(_ int, row *goquery.Selection) {
			cols := row.Find("td")
			if cols.Length() == 2 {
				// Get element name
				elementName, exists := cols.Eq(0).Find("a").Attr("title")
				if !exists || elementName == "" {
					elementName = strings.TrimSpace(cols.Eq(0).Find("a").Text())
				}

				if elementName == "" {
					return
				}

				// Add this element to the current table's elements
				elementsInThisTable = append(elementsInThisTable, elementName)

				// Extract recipes for this element
				elementRecipes := [][]string{}
				cols.Eq(1).Find("li").Each(func(_ int, li *goquery.Selection) {
					recipe := []string{}
					li.Find("a").Each(func(_ int, a *goquery.Selection) {
						text := a.Text()
						if text == "" {
							return
						}

						ingredient, exists := a.Attr("title")
						if !exists || ingredient == "" {
							ingredient = strings.TrimSpace(text)
						}
						if ingredient != "" {
							recipe = append(recipe, ingredient)
						}
					})
					if len(recipe) > 0 {
						elementRecipes = append(elementRecipes, recipe)
					}
				})

				if len(elementRecipes) > 0 {
					recipesInThisTable[elementName] = elementRecipes
				}
			}
		})

		// Store elements found in this table
		elementsPerTable[i+1] = elementsInThisTable

		// Copy recipes from previous table
		allRecipesUpToTable[i+1] = make(map[string][][]string)
		for elem, recipes := range allRecipesUpToTable[i] {
			allRecipesUpToTable[i+1][elem] = recipes
		}

		// Add new recipes
		for elem, recipes := range recipesInThisTable {
			allRecipesUpToTable[i+1][elem] = recipes
		}

		// Update available elements
		for _, element := range elementsInThisTable {
			if !allElements[element] {
				result.Elements = append(result.Elements, element)
				allElements[element] = true
			}
		}
	}

	// Create a map to track which elements we've already processed
	// processedElements := make(map[string]bool)

	// For this step, we'll build the initial recipe structure first (just like in the original code)
	// Then we'll add the recursive recipes in a separate pass
	intermediateResult := OutputData{
		Elements: result.Elements,
		Recipes:  make(map[string]map[string][][]string),
	}

	// Step 3: Build the initial result structure exactly as in the original code
	for tableIndex := 1; tableIndex <= tableCount; tableIndex++ {
		// Get available elements up to this table
		availableElements := map[string]bool{
			"Air":   true,
			"Earth": true,
			"Fire":  true,
			"Water": true,
		}

		for i := 1; i <= tableIndex; i++ {
			for _, element := range elementsPerTable[i] {
				availableElements[element] = true
			}
		}

		// Process new elements from this table
		currentTableElements := elementsPerTable[tableIndex]
		for _, element := range currentTableElements {
			recipes, exists := allRecipesUpToTable[tableIndex][element]
			if !exists {
				continue
			}

			// Only include valid recipes (using available elements)
			validRecipes := [][]string{}
			for _, recipe := range recipes {
				valid := true
				for _, ingredient := range recipe {
					if !availableElements[ingredient] {
						valid = false
						break
					}
				}
				if valid {
					validRecipes = append(validRecipes, recipe)
				}
			}

			if len(validRecipes) > 0 {
				// Initialize component map for this element
				if intermediateResult.Recipes[element] == nil {
					intermediateResult.Recipes[element] = make(map[string][][]string)
				}

				// Add the element's own recipes
				intermediateResult.Recipes[element][element] = validRecipes

				// Add component recipes (only one level, as in the original code)
				for _, recipe := range validRecipes {
					for _, ingredient := range recipe {
						if ingredient == element {
							continue // Skip self-reference
						}

						// Only use ingredients that are available up to this table
						if !availableElements[ingredient] {
							continue
						}

						ingredientRecipes, exists := allRecipesUpToTable[tableIndex][ingredient]
						if !exists || len(ingredientRecipes) == 0 {
							continue
						}

						// Filter ingredient recipes to only use available elements
						validIngredientRecipes := [][]string{}
						for _, ingRecipe := range ingredientRecipes {
							valid := true
							for _, ing := range ingRecipe {
								if !availableElements[ing] || ing == element {
									valid = false
									break
								}
							}
							if valid {
								validIngredientRecipes = append(validIngredientRecipes, ingRecipe)
							}
						}

						if len(validIngredientRecipes) > 0 {
							intermediateResult.Recipes[element][ingredient] = validIngredientRecipes
						}
					}
				}
			}
		}
	}

	// Step 4: Now build the final result with recursive recipes
	// We'll use the intermediate result as a starting point

	// The base elements don't need recipes
	baseElements := map[string]bool{
		"Air":   true,
		"Earth": true,
		"Fire":  true,
		"Water": true,
	}

	// Copy the elements list
	result.Elements = intermediateResult.Elements

	// For each element in the intermediate result
	for element, elementRecipes := range intermediateResult.Recipes {
		// Skip base elements
		if baseElements[element] {
			continue
		}

		// Initialize the recipe map for this element in the final result
		result.Recipes[element] = make(map[string][][]string)

		// First, copy the element's own recipes
		result.Recipes[element][element] = elementRecipes[element]

		// Process each ingredient
		for _, recipe := range elementRecipes[element] {
			for _, ingredient := range recipe {
				// Skip base elements
				if baseElements[ingredient] {
					continue
				}

				// Reset processed elements for each new top-level ingredient
				localProcessed := make(map[string]bool)
				localProcessed[element] = true // Avoid recursion back to the target element

				// Recursively get recipes for this ingredient
				getRecursiveRecipes(
					ingredient,
					ingredient,
					intermediateResult.Recipes,
					result.Recipes[element],
					localProcessed,
					baseElements,
				)
			}
		}
	}

	return result, nil
}

// getRecursiveRecipes recursively adds recipes for an ingredient and its components
func getRecursiveRecipes(
	currentElement string,
	originalIngredient string,
	allRecipes map[string]map[string][][]string,
	targetRecipes map[string][][]string,
	processed map[string]bool,
	baseElements map[string]bool,
) {
	// Mark as processed to avoid cycles
	processed[currentElement] = true

	// Get recipes for the current element
	elementRecipes, exists := allRecipes[currentElement]
	if !exists {
		return
	}

	// Add the direct recipes for this element if it's the original ingredient or its subcomponent
	if directRecipes, ok := elementRecipes[currentElement]; ok && len(directRecipes) > 0 {
		// Only add if not already present
		if _, exists := targetRecipes[currentElement]; !exists {
			targetRecipes[currentElement] = directRecipes
		}
	}

	// Get the component recipes if this is the original ingredient
	if currentElement == originalIngredient {
		// Add component recipes for all ingredients in recipes of currentElement
		for ingredient, recipes := range elementRecipes {
			if ingredient == currentElement || baseElements[ingredient] || processed[ingredient] {
				continue
			}

			// Add these recipes
			targetRecipes[ingredient] = recipes

			// Recursively process this ingredient
			getRecursiveRecipes(
				ingredient,
				originalIngredient,
				allRecipes,
				targetRecipes,
				processed,
				baseElements,
			)
		}
	}

	// For each recipe of the current element, process its ingredients recursively
	selfRecipes, exists := elementRecipes[currentElement]
	if exists {
		for _, recipe := range selfRecipes {
			for _, ingredient := range recipe {
				// Skip base elements and already processed elements
				if baseElements[ingredient] || processed[ingredient] {
					continue
				}

				// Get recipes for this ingredient
				ingredientRecipes, exists := allRecipes[ingredient]
				if !exists || len(ingredientRecipes) == 0 {
					continue
				}

				// If this ingredient has its own recipes, add them
				if directRecipes, ok := ingredientRecipes[ingredient]; ok && len(directRecipes) > 0 {
					// Only add if not already present
					if _, exists := targetRecipes[ingredient]; !exists {
						targetRecipes[ingredient] = directRecipes
					}
				}

				// Recursively process this ingredient
				getRecursiveRecipes(
					ingredient,
					originalIngredient,
					allRecipes,
					targetRecipes,
					processed,
					baseElements,
				)
			}
		}
	}
}

// SaveRecipesToJson saves the recipes to a JSON file.
func SaveRecipesToJson(data OutputData, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
