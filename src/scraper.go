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

// SimpleOutputData is a simplified format for the recipe data
type SimpleOutputData struct {
	Elements []string   `json:"elements"` // List of all element names
	Recipes  RecipesMap `json:"recipes"`  // Map of element name to its recipes
}

// RecipesMap maps element names to their recipes
type RecipesMap map[string][][]string

func ScrapeSimpleRecipes() (SimpleOutputData, error) {
	// Initialize the result structure with starting elements
	result := SimpleOutputData{
		Elements: []string{"Air", "Earth", "Fire", "Water"},
		Recipes:  make(map[string][][]string),
	}

	// Define the elements to exclude
	excludedElements := map[string]bool{
		"Time":         true,
		"Ruins":        true,
		"Archeologist": true,
	}

	res, err := http.Get(url)
	if err != nil {
		return SimpleOutputData{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return SimpleOutputData{}, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return SimpleOutputData{}, err
	}

	// Keep track of unique elements
	allElements := map[string]bool{
		"Air":   true,
		"Earth": true,
		"Fire":  true,
		"Water": true,
	}

	// Find all tables that follow a paragraph with our target text
	doc.Find("p").Each(func(_ int, p *goquery.Selection) {
		if strings.Contains(p.Text(), "These elements can be created by combining only") {
			nextTable := p.NextFiltered("table")
			if nextTable.Length() > 0 {
				// Extract data from this table
				rows := nextTable.Find("tr").Slice(1, nextTable.Find("tr").Length())
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

						// Skip excluded elements
						if excludedElements[elementName] {
							fmt.Printf("Skipping excluded element: %s", elementName)
							return
						}

						// Add element to our overall elements list if it's new
						if !allElements[elementName] {
							result.Elements = append(result.Elements, elementName)
							allElements[elementName] = true
						}

						// Extract recipes for this element
						elementRecipes := [][]string{}
						cols.Eq(1).Find("li").Each(func(_ int, li *goquery.Selection) {
							recipe := []string{}
							containsExcluded := false

							li.Find("a").Each(func(_ int, a *goquery.Selection) {
								text := a.Text()
								if text == "" {
									return
								}

								ingredient, exists := a.Attr("title")
								if !exists || ingredient == "" {
									ingredient = strings.TrimSpace(text)
								}

								// Check if this is an excluded ingredient
								if excludedElements[ingredient] {
									containsExcluded = true
									return
								}

								if ingredient != "" {
									recipe = append(recipe, ingredient)
								}
							})

							// Only add recipes that don't contain excluded elements
							if len(recipe) > 0 && !containsExcluded {
								elementRecipes = append(elementRecipes, recipe)
							}
						})

						if len(elementRecipes) > 0 {
							result.Recipes[elementName] = elementRecipes
							fmt.Printf("Added %d recipes for element: %s", len(elementRecipes), elementName)
						} else {
							fmt.Printf("No valid recipes found for element: %s", elementName)
						}
					}
				})
			}
		}
	})

	// fmt the result count for debugging
	fmt.Printf("Total elements: %d, Total elements with recipes: %d",
		len(result.Elements), len(result.Recipes))

	return result, nil
}

// ScrapeInitialRecipes scrapes recipes progressively from the Little Alchemy 2 wiki.
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

	// Step 2: Process tables one by one, building up available elements
	allElements := map[string]bool{
		"Air":   true,
		"Earth": true,
		"Fire":  true,
		"Water": true,
	}

	// This will store all recipes regardless of table/tier
	allPossibleRecipes := make(map[string][][]string)

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
					// Store in our all-possible-recipes map
					allPossibleRecipes[elementName] = elementRecipes
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

	// Create a map to keep track of elements by tier
	elementTiers := make(map[string]int)
	// Base elements are tier 0
	elementTiers["Air"] = 0
	elementTiers["Earth"] = 0
	elementTiers["Fire"] = 0
	elementTiers["Water"] = 0

	// Assign tiers to each element
	for tier := 1; tier <= tableCount; tier++ {
		for _, element := range elementsPerTable[tier] {
			elementTiers[element] = tier
		}
	}

	// The base elements don't need recipes
	baseElements := map[string]bool{
		"Air":   true,
		"Earth": true,
		"Fire":  true,
		"Water": true,
	}

	// Final step: Build the result structure with recursive recipes
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
				if result.Recipes[element] == nil {
					result.Recipes[element] = make(map[string][][]string)
				}

				// Add the element's own recipes
				result.Recipes[element][element] = validRecipes

				// Now let's gather recursive recipes
				for _, recipe := range validRecipes {
					for _, ingredient := range recipe {
						// Skip base elements and self-references
						if baseElements[ingredient] || ingredient == element {
							continue
						}

						// Create a map to track processed elements (to avoid cycles)
						processedElements := make(map[string]bool)
						processedElements[element] = true // Don't recurse back to the parent

						// Get recursive recipes
						collectRecursiveRecipes(
							ingredient,
							ingredient,
							allPossibleRecipes,
							result.Recipes[element],
							processedElements,
							baseElements,
							availableElements,
							elementTiers,
							tableIndex,
						)
					}
				}
			}
		}
	}

	return result, nil
}

// collectRecursiveRecipes gathers all valid recipes for an ingredient recursively
func collectRecursiveRecipes(
	currentElement string,
	originalIngredient string,
	allRecipes map[string][][]string,
	targetRecipes map[string][][]string,
	processed map[string]bool,
	baseElements map[string]bool,
	availableElements map[string]bool,
	elementTiers map[string]int,
	currentTableTier int,
) {
	// Mark this element as processed to avoid cycles
	processed[currentElement] = true

	// Get recipes for the current element
	elementRecipes, exists := allRecipes[currentElement]
	if !exists {
		return
	}

	// Only add recipes using elements that are available at current tier
	validRecipes := [][]string{}
	for _, recipe := range elementRecipes {
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

	// Add to target recipes if we have valid recipes
	if len(validRecipes) > 0 && currentElement == originalIngredient {
		targetRecipes[currentElement] = validRecipes
	}

	// For each recipe of the current element, process its ingredients recursively
	for _, recipe := range validRecipes {
		for _, ingredient := range recipe {
			// Skip base elements and already processed elements
			if baseElements[ingredient] || processed[ingredient] {
				continue
			}

			// Get recipes for this sub-ingredient
			subRecipes, exists := allRecipes[ingredient]
			if !exists {
				continue
			}

			// Filter valid recipes
			validSubRecipes := [][]string{}
			for _, subRecipe := range subRecipes {
				valid := true
				for _, subIng := range subRecipe {
					if !availableElements[subIng] {
						valid = false
						break
					}
				}
				if valid {
					validSubRecipes = append(validSubRecipes, subRecipe)
				}
			}

			// Add to target if we have valid recipes
			if len(validSubRecipes) > 0 {
				targetRecipes[ingredient] = validSubRecipes
			}

			// Recursively process this ingredient
			collectRecursiveRecipes(
				ingredient,
				originalIngredient,
				allRecipes,
				targetRecipes,
				processed,
				baseElements,
				availableElements,
				elementTiers,
				currentTableTier,
			)
		}
	}
}

// ScrapeRecipes scrapes recipe data filtering by already available elements
// ScrapeRecipes scrapes recipe data with consistent sub-recipes
// ScrapeRecipes scrapes recipe data with recursive consistent sub-recipes
func ScrapeRecipes() (OutputData, error) {
	// Initialize the result structure with starting elements
	result := OutputData{
		Elements: []string{"Air", "Earth", "Fire", "Water"},
		Recipes:  make(map[string]map[string][][]string),
	}

	// Define base elements
	baseElements := map[string]bool{
		"Air":   true,
		"Earth": true,
		"Fire":  true,
		"Water": true,
	}

	// Define the elements to exclude
	excludedElements := map[string]bool{
		"Time":         true,
		"Ruins":        true,
		"Archeologist": true,
	}

	// Track available elements (initially just base elements)
	availableElements := map[string]bool{
		"Air":   true,
		"Earth": true,
		"Fire":  true,
		"Water": true,
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

	// Store all recipes we find
	allRecipes := make(map[string][][]string)

	// Process tables in order (which handles tiers implicitly)
	doc.Find("p").Each(func(_ int, p *goquery.Selection) {
		if strings.Contains(p.Text(), "These elements can be created by combining only") {
			nextTable := p.NextFiltered("table")
			if nextTable.Length() > 0 {
				// Elements in this table/tier
				elementsInCurrentTable := []string{}

				// Extract data from this table
				rows := nextTable.Find("tr").Slice(1, nextTable.Find("tr").Length())
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

						// Skip excluded elements
						if excludedElements[elementName] {
							fmt.Printf("Skipping excluded element: %s\n", elementName)
							return
						}

						// Track this element for this table
						elementsInCurrentTable = append(elementsInCurrentTable, elementName)

						// Extract all recipes for this element
						elementRecipes := [][]string{}
						cols.Eq(1).Find("li").Each(func(_ int, li *goquery.Selection) {
							recipe := []string{}
							containsExcluded := false

							li.Find("a").Each(func(_ int, a *goquery.Selection) {
								text := a.Text()
								if text == "" {
									return
								}

								ingredient, exists := a.Attr("title")
								if !exists || ingredient == "" {
									ingredient = strings.TrimSpace(text)
								}

								// Check if this is an excluded ingredient
								if excludedElements[ingredient] {
									containsExcluded = true
									return
								}

								if ingredient != "" {
									recipe = append(recipe, ingredient)
								}
							})

							// Only add recipes that don't contain excluded elements
							if len(recipe) > 0 && !containsExcluded {
								elementRecipes = append(elementRecipes, recipe)
							}
						})

						if len(elementRecipes) > 0 {
							allRecipes[elementName] = elementRecipes
						}
					}
				})

				// Now process elements in this table
				for _, element := range elementsInCurrentTable {
					recipes, exists := allRecipes[element]
					if !exists || len(recipes) == 0 {
						continue
					}

					// Filter recipes to only include those with available ingredients
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

					// If we have valid recipes, add this element to result
					if len(validRecipes) > 0 {
						// Add to elements list and mark as available
						result.Elements = append(result.Elements, element)
						availableElements[element] = true

						// Initialize recipe map for this element
						result.Recipes[element] = make(map[string][][]string)

						// Add the valid recipes for this element
						result.Recipes[element][element] = validRecipes

						// Now recursively add recipes for all non-base ingredients
						addedIngredients := make(map[string]bool) // Avoid duplicates
						addRecipesRecursively(element, validRecipes, result, baseElements, addedIngredients)

						fmt.Printf("Added element %s with %d valid recipes\n", element, len(validRecipes))
					}
				}
			}
		}
	})

	fmt.Printf("Total elements: %d, Total elements with recipes: %d\n",
		len(result.Elements), len(result.Recipes))

	return result, nil
}

// addRecipesRecursively adds recipes for all non-base ingredients recursively
func addRecipesRecursively(targetElement string, recipes [][]string, result OutputData,
	baseElements map[string]bool, addedIngredients map[string]bool) {
	for _, recipe := range recipes {
		for _, ingredient := range recipe {
			// Skip base elements and already processed ingredients for this target
			if baseElements[ingredient] || addedIngredients[ingredient] {
				continue
			}

			// Mark this ingredient as processed for this target
			addedIngredients[ingredient] = true

			// Find this ingredient's recipes in the result
			if ingredientRecipesMap, exists := result.Recipes[ingredient]; exists {
				if ingredientRecipes, exists := ingredientRecipesMap[ingredient]; exists {
					// Add this ingredient's recipes to the target element
					result.Recipes[targetElement][ingredient] = ingredientRecipes

					// Recursively add this ingredient's ingredients' recipes
					addRecipesRecursively(targetElement, ingredientRecipes, result, baseElements, addedIngredients)
				}
			}
		}
	}
}

// func main() {
// 	// Use your new scraping function here
// 	data, err := ScrapeRecipes()
// 	if err != nil {
// 		fmt.Printf("Error scraping recipes: %v\n", err)
// 		return
// 	}

// 	// Save to recipes.json
// 	err = SaveRecipesToJson(data, "recipes.json")
// 	if err != nil {
// 		fmt.Printf("Error saving recipes: %v\n", err)
// 		return
// 	}

// 	fmt.Println("Recipe data saved to recipes.json successfully!")
// }

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

func SaveSimpleRecipesToJson(data SimpleOutputData, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully saved %d elements with recipes to %s\n", len(data.Elements), filename)
	return nil
}
