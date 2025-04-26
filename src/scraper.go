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

// ScrapeInitialRecipes scrapes the initial recipes from the Little Alchemy 2 wiki.
func ScrapeInitialRecipes() (OutputData, error) {
	// Initialize the main result structure
	result := OutputData{
		Elements: []string{"Air", "Earth", "Fire", "Water"},
		Recipes:  make(map[string]map[string][][]string),
	}

	// Keep a flat map of all recipes for easier processing
	allRecipes := make(map[string][][]string)

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

	// Cari elemen p yang berisi teks target
	targetParagraph := doc.Find("p").FilterFunction(func(_ int, s *goquery.Selection) bool {
		return strings.Contains(s.Text(), "These elements can be created by combining only the ")
	}).First()

	// Cari tabel yang berada setelah elemen p tersebut
	table := targetParagraph.NextFiltered("table")

	// Temukan semua baris tabel (tr)
	rows := table.Find("tr").Slice(1, table.Find("tr").Length())

	// First pass: collect all recipes and element names
	rows.Each(func(_ int, row *goquery.Selection) {
		cols := row.Find("td")
		if cols.Length() == 2 {
			// Ambil nama elemen dari title td pertama
			elementName, exists := cols.Eq(0).Find("a").Attr("title")
			if !exists || elementName == "" {
				elementName = strings.TrimSpace(cols.Eq(0).Find("a").Text())
			}

			if elementName == "" {
				return // Skip if we couldn't get the element name
			}

			// Add to the Elements array, but avoid duplicates
			found := false
			for _, el := range result.Elements {
				if el == elementName {
					found = true
					break
				}
			}
			if !found {
				result.Elements = append(result.Elements, elementName)
			}

			// Initialize recipe array if it doesn't exist
			if _, exists := allRecipes[elementName]; !exists {
				allRecipes[elementName] = [][]string{}
			}

			// Ambil resep dari td kedua
			cols.Eq(1).Find("li").Each(func(_ int, li *goquery.Selection) {
				recipe := []string{}
				li.Find("a").Each(func(_ int, a *goquery.Selection) {
					text := a.Text()
					if text == "" {
						return // Skip empty elements
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
					allRecipes[elementName] = append(allRecipes[elementName], recipe)
				}
			})
		}
	})

	// Second pass: build the structured result
	result.Recipes = make(map[string]map[string][][]string)

	for element, recipes := range allRecipes {
		componentMap := make(map[string][][]string)

		// Add the element's own recipes
		componentMap[element] = recipes

		// Add component recipes if available
		for _, recipe := range recipes {
			for _, ingredient := range recipe {
				if ingredientRecipes, exists := allRecipes[ingredient]; exists && len(ingredientRecipes) > 0 {
					componentMap[ingredient] = ingredientRecipes
				}
			}
		}

		// Filter recipes and add to result
		filteredComponentMap := make(map[string][][]string)
		for comp, compRecipes := range componentMap {
			filteredRecipes := [][]string{}
			for _, recipe := range compRecipes {
				valid := true
				for _, ing := range recipe {
					found := false
					for _, el := range result.Elements {
						if el == ing {
							found = true
							break
						}
					}
					if !found || ing == element { // Avoid self-reference
						valid = false
						break
					}
				}
				if valid {
					filteredRecipes = append(filteredRecipes, recipe)
				}
			}
			if len(filteredRecipes) > 0 {
				filteredComponentMap[comp] = filteredRecipes
			}
		}

		if len(filteredComponentMap) > 0 {
			result.Recipes[element] = filteredComponentMap
		}
	}

	return result, nil
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

// func main() {
//  data, err := ScrapeInitialRecipes()
//  if err != nil {
//      log.Fatalf("Error scraping recipes: %v", err)
//  }

//  err = SaveRecipesToJson(data, "initial_recipes.json")
//  if err != nil {
//      log.Fatalf("Error saving recipes to JSON: %v", err)
//  }

//  fmt.Println("Initial recipes scraped and saved to initial_recipes.json")
// }
