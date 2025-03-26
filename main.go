package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Product struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Price    string `json:"price"`
	Image    string `json:"image"`
}

type APIResponse struct {
	PageData struct {
		Products []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Prices struct {
				CurrentPrice float64 `json:"currentPrice"`
			} `json:"prices"`
			Media struct {
				Full []struct {
					Src string `json:"src"`
				} `json:"full"`
			} `json:"media"`
		} `json:"products"`
	} `json:"pageData"`
}

func downloadImage(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download image %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-200 response code %d for image %s", resp.StatusCode, url)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filepath, err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save image %s: %v", url, err)
	}

	return nil
}

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

func main() {

	testMode := flag.Bool("test", false, "Enable test mode: scrape limited data")
	flag.BoolVar(testMode, "t", false, "Enable test mode: scrape limited data")
	flag.Parse()

	categories := []string{"handbags", "wallets", "jewelry", "shoes", "clothing", "accessories"}
	excludedCategories := map[string]bool{
		"new":   true,
		"home":  true,
		"gifts": true,
		"sale":  true,
	}

	var allProducts []Product

	imageDir := "data/images"
	ensureDir(imageDir)

	for _, category := range categories {
		if excludedCategories[category] {
			fmt.Printf("Skipping category: %s\n", category)
			continue
		}

		fmt.Printf("Scraping category: %s\n", category)
		page := 1

		maxPages := 1
		if !*testMode {
			maxPages = 1000
		}
		for ; page <= maxPages; page++{
			apiURL := fmt.Sprintf("https://www.katespade.com/api/get-shop/%s/view-all?page=%d", category, page)
			fmt.Printf("Fetching API URL: %s\n", apiURL)

			resp, err := http.Get(apiURL)
			if err != nil {
				log.Printf("Failed to fetch API: %v", err)
				break
			}
			defer resp.Body.Close()

			var apiResp APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				log.Printf("Failed to parse JSON response: %v", err)
				break
			}

			if len(apiResp.PageData.Products) == 0 {
				fmt.Printf("No more products found in category: %s\n", category)
				break
			}

			for _, product := range apiResp.PageData.Products {
				price := fmt.Sprintf("$%.2f", product.Prices.CurrentPrice)
				imageURL := ""
				if len(product.Media.Full) > 0 {
					imageURL = product.Media.Full[0].Src
				}

				imageFileName := fmt.Sprintf("%s.jpg", product.ID)
				imagePath := filepath.Join(imageDir, imageFileName)

				if imageURL != "" {
					err := downloadImage(imageURL, imagePath)
					if err != nil {
						log.Printf("Error downloading image for product %s: %v", product.ID, err)
					}
				}

				prod := Product{
					ID:       product.ID,
					Title:    product.Name,
					Category: category,
					Price:    price,
					Image:    imageFileName,
				}

				allProducts = append(allProducts, prod)
				fmt.Printf("ID: %s\nTitle: %s\nCategory: %s\nPrice: %s\nImage: %s\n---------------------------\n",
					prod.ID, prod.Title, prod.Category, prod.Price, prod.Image)
			}

			if *testMode {
				break
			}
		}
	}

	file, err := json.MarshalIndent(allProducts, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal products: %v", err)
	}

	if err := os.WriteFile("data/products.json", file, 0644); err != nil {
		log.Fatalf("Failed to write JSON file: %v", err)
	}

	fmt.Println("Scraping complete! Data saved to products.json")
}
