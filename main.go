package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Product struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Category string `json:"category"`
	Price    string `json:"price"`
	Image    string `json:"image"`
}

func fetchCategoryData(category string) []Product {
	apiURL := fmt.Sprintf("https://www.katespade.com/api/get-shop/%s/view-all?page=1", category)
	fmt.Printf("Fetching API URL: %s\n", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		log.Fatalf("Failed to fetch API: %v", err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		log.Fatalf("Failed to decode API response: %v", err)
	}

	pageData := result["pageData"].(map[string]interface{})
	productsData := pageData["products"].([]interface{})
	breadcrumbs := pageData["breadcrumbs"].([]interface{})

	categoryName := ""
	if len(breadcrumbs) > 0 {
		categoryName = breadcrumbs[0].(map[string]interface{})["htmlValue"].(string)
	}

	var products []Product
	for _, p := range productsData {
		product := p.(map[string]interface{})

		id := product["productId"].(string)
		title := product["name"].(string)

		// Extract price
		price := ""
		if defaultVariant, ok := product["defaultVariant"].(map[string]interface{}); ok {
			if prices, ok := defaultVariant["prices"].(map[string]interface{}); ok {
				if currentPrice, ok := prices["currentPrice"].(float64); ok {
					price = fmt.Sprintf("$%.2f", currentPrice)
				}
			}
		}

		// Extract image
		image := ""
		if defaultColor, ok := product["defaultColor"].(map[string]interface{}); ok {
			if img, ok := defaultColor["image"].(map[string]interface{}); ok {
				image = img["src"].(string)
			}
		}

		products = append(products, Product{
			ID:       id,
			Title:    title,
			Category: categoryName,
			Price:    price,
			Image:    image,
		})
	}

	return products
}

func saveToFile(products []Product) {
	file, err := os.Create("products.json")
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(products)
	if err != nil {
		log.Fatalf("Failed to write JSON to file: %v", err)
	}

	fmt.Println("Scraping complete! Data saved to products.json")
}

func main() {
	category := "handbags" // Change this to the desired category
	products := fetchCategoryData(category)

	for _, p := range products {
		fmt.Printf("ID: %s\nTitle: %s\nCategory: %s\nPrice: %s\nImage: %s\n---------------------------\n", p.ID, p.Title, p.Category, p.Price, p.Image)
	}

	saveToFile(products)
}
