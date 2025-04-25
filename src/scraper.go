package main

import (
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// scraper fetches and parses the wiki page into the global `recipes` map.
func scraper() (map[string][][]string, error) {
	url := "https://little-alchemy.fandom.com/wiki/Elements_(Little_Alchemy_2)"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	// Temporary map
	result := make(map[string][][]string)

	// Example selector — sesuaikan dengan struktur resep di halaman
	doc.Find(".recipe-card").Each(func(_ int, s *goquery.Selection) {
		name := s.Find(".element-name").Text()
		s.Find(".combination img").Each(func(_ int, img *goquery.Selection) {
			if title, ok := img.Attr("title"); ok {
				// ...
			}
		})
		// Isi result[name] = append(...)
	})

	return result, nil
}
