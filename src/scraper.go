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
