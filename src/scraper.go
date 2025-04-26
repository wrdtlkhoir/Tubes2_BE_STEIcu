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

// Result structure will represent your desired output format
type Result map[string]map[string][][]string

// ScrapeInitialRecipes scrapes the initial recipes from the Little Alchemy 2 wiki.
func ScrapeInitialRecipes() (Result, error) {
	// Initialize the main result structure
	result := make(Result)
	// Also keep a flat map of all recipes for easier processing
	allRecipes := make(map[string][][]string)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	fmt.Println("Successfully fetched the URL")

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	// Cari elemen p yang berisi teks target
	targetParagraph := doc.Find("p").FilterFunction(func(_ int, s *goquery.Selection) bool {
		return strings.Contains(s.Text(), "These elements can be created by combining only the ")
	}).First()

	// Debug: Print the text of the found paragraph
	fmt.Println("Target paragraph found:", targetParagraph.Text())

	// Cari tabel yang berada setelah elemen p tersebut
	table := targetParagraph.NextFiltered("table")

	// Debug: Check if table was found
	if table.Length() == 0 {
		fmt.Println("Table not found after target paragraph!")

		// Print all paragraphs to help troubleshoot
		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			fmt.Printf("Paragraph %d: %s\n", i, s.Text())
		})
	} else {
		fmt.Printf("Found table with %d rows\n", table.Find("tr").Length())
	}

	// Temukan semua baris tabel (tr)
	rows := table.Find("tr").Slice(1, table.Find("tr").Length()) // Lewati baris header

	// Debug: Record the number of elements we find
	elementCount := 0

	// First pass: collect all recipes
	rows.Each(func(i int, row *goquery.Selection) {
		cols := row.Find("td")
		if cols.Length() == 2 {
			// Updated selector for element name
			elementName, exists := cols.Eq(0).Find("a").Attr("title")
			if !exists || elementName == "" {
				fmt.Printf("Row %d: Could not find element name title attribute\n", i+1)
				// Try different approach
				elementName = strings.TrimSpace(cols.Eq(0).Find("a").Text())
				fmt.Printf("Using text instead: '%s'\n", elementName)
			}

			if elementName == "" {
				fmt.Printf("Row %d: Skipping empty element name\n", i+1)
				return // Skip if we couldn't get the element name
			}

			elementCount++

			// Initialize recipe array if it doesn't exist
			if _, exists := allRecipes[elementName]; !exists {
				allRecipes[elementName] = [][]string{}
			}

			// Ambil resep dari td kedua
			recipeCount := 0
			cols.Eq(1).Find("li").Each(func(j int, li *goquery.Selection) {
				recipe := []string{}
				li.Find("a").Each(func(k int, a *goquery.Selection) {
					ingredient, exists := a.Attr("title")
					if !exists || ingredient == "" {
						fmt.Printf("Row %d, Recipe %d: Missing ingredient title\n", i+1, j+1)
						ingredient = strings.TrimSpace(a.Text())
						fmt.Printf("Using text instead: '%s'\n", ingredient)
					}
					if ingredient != "" {
						recipe = append(recipe, ingredient)
					}
				})
				if len(recipe) > 0 {
					allRecipes[elementName] = append(allRecipes[elementName], recipe)
					recipeCount++
				}
			})
			fmt.Printf("Element '%s' has %d recipes\n", elementName, recipeCount)
		} else {
			fmt.Printf("Row %d: Expected 2 columns, found %d\n", i+1, cols.Length())
		}
	})

	fmt.Printf("First pass: Found %d elements with recipes\n", elementCount)

	// Debug: Print first few recipes
	count := 0
	for element, recipes := range allRecipes {
		if count < 5 {
			fmt.Printf("Element: %s, Recipes: %v\n", element, recipes)
		}
		count++
	}

	// Second pass: build the structured result
	for element, recipes := range allRecipes {
		componentMap := make(map[string][][]string)

		// Add the element's own recipes
		componentMap[element] = recipes

		// Add component recipes (just the first row for now as requested)
		for _, recipe := range recipes {
			for _, ingredient := range recipe {
				// Only add non-basic elements (those that have recipes)
				if ingredientRecipes, exists := allRecipes[ingredient]; exists && len(ingredientRecipes) > 0 {
					// Just add the first recipe row
					if len(ingredientRecipes) > 0 {
						componentMap[ingredient] = [][]string{ingredientRecipes[0]}
					}
				}
			}
		}

		// Add to result
		result[element] = componentMap
	}

	fmt.Printf("Final result has %d elements\n", len(result))

	// Debug: Print a sample of the result
	count = 0
	for element, components := range result {
		if count < 3 {
			fmt.Printf("Element: %s\n", element)
			for comp, recipes := range components {
				fmt.Printf("  Component: %s, Recipes: %v\n", comp, recipes)
			}
		}
		count++
	}

	return result, nil
}

// SaveRecipesToJson saves the recipes to a JSON file.
func SaveRecipesToJson(recipes Result, filename string) error {
	jsonData, err := json.MarshalIndent(recipes, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

// func main() {
// 	recipes, err := ScrapeInitialRecipes()
// 	if err != nil {
// 		log.Fatalf("Error scraping recipes: %v", err)
// 	}

// 	err = SaveRecipesToJson(recipes, "initial_recipes.json")
// 	if err != nil {
// 		log.Fatalf("Error saving recipes to JSON: %v", err)
// 	}

// 	fmt.Println("Initial recipes scraped and saved to initial_recipes.json")
// }
