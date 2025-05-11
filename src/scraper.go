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
