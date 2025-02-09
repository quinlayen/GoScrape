package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/queue"
)

const baseURL = "https://www.katespade.com"

// Product struct to store product details
type Product struct {
	ID           string   `json:"item-id"`
	Title        string   `json:"item-name"`
	Price        string   `json:"item-price"`
	Measurements []string `json:"item-measurements,omitempty"`
	Materials    []string `json:"item-materials,omitempty"`
	Features     []string `json:"item-features,omitempty"`
	Img          string   `json:"img,omitempty"`
	Category     string   `json:"category,omitempty"`
}

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.katespade.com", "katespade.com"),
	)

	q, err := queue.New(2, &queue.InMemoryQueueStorage{MaxSize: 10000})
	if err != nil {
		log.Fatalf("Failed to initialize queue: %v", err)
	}

	categoryLinks := getCategoryLinks(c)
	for _, category := range categoryLinks {
		if shouldSkipCategory(category) {
			fmt.Println("Skipping category:", category)
			continue
		}
		fullURL := baseURL + category
		err := q.AddURL(fullURL)
		if err != nil {
			fmt.Println("Failed to add category link:", err)
		}
	}

	var products []Product

	productCollector := colly.NewCollector(
		colly.AllowedDomains("www.katespade.com", "katespade.com"),
	)

	// Debugging: Confirm product pages are being visited
	productCollector.OnRequest(func(r *colly.Request) {
		//r.Ctx.Put("category", extractCategoryFromURL(r.URL.String()))
		fmt.Println("Visiting product page:", r.URL.String())
	})

	// Handle errors during product page visits
	productCollector.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Request to %s failed with status %d: %v\n", r.Request.URL, r.StatusCode, err)
	})

	scrapeProducts(productCollector, q, &products)

	err = exportToJSON(products, "products.json")
	if err != nil {
		log.Fatalf("Failed to export to JSON: %v", err)
	}

	fmt.Println("Scraping complete! Data saved to products.json")
} // End of main

// getCategoryLinks scrapes the main page to extract category links
func getCategoryLinks(c *colly.Collector) []string {
	var categories []string

	c.OnHTML("div.menu-tier-1.css-1h5bbey div.css-wnawyw div.css-19rc50l a", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if strings.HasPrefix(href, "/shop/") && strings.Contains(href, "view-all") {
			categories = append(categories, href)
			fmt.Println("Found category link:", href)
		}
	})

	err := c.Visit(baseURL)
	if err != nil {
		log.Fatal("Failed to visit main page:", err)
	}
	return categories
}

// shouldSkipCategory checks if a category should be skipped
func shouldSkipCategory(category string) bool {
	skipCategories := []string{"/shop/new/view-all", "/shop/home/view-all", "/shop/gifts/view-all", "/shop/sale/view-all"}
	for _, skip := range skipCategories {
		if strings.Contains(category, skip) {
			return true
		}
	}
	return false
}

// scrapeProducts handles scraping product listings and individual product details
func scrapeProducts(productCollector *colly.Collector, q *queue.Queue, products *[]Product) {

	productCollector.OnHTML("div.product-tile", func(e *colly.HTMLElement) {
		productLink := e.ChildAttr("div.product-name a", "href")
		if productLink != "" {
			fullProductURL := baseURL + productLink

			// Extract category from the current category page URL
			category := extractCategoryFromURL(e.Request.URL.String())

			fmt.Println("Found product link:", fullProductURL, "in category:", category)

			// Create a new request with the category context
			ctx := colly.NewContext()
			ctx.Put("category", category)
			//err := e.Request.Ctx.Clone().Put("category", category)
			//if err != nil {
			//	fmt.Println("Failed to set category context:", err)
			//}

			// Visit the product page with the category context attached
			e.Request.Visit(fullProductURL, ctx)
		}
	})

	// Enable pagination
	productCollector.OnHTML("a.pagination-next", func(e *colly.HTMLElement) {
		nextPage := e.Request.AbsoluteURL(e.Attr("href"))
		fmt.Println("Found next page:", nextPage)
		err := q.AddURL(nextPage)
		if err != nil {
			fmt.Println("Failed to add next page:", err)
		}
	})

	// Corrected XPath selectors for product details
	productCollector.OnXML("//div[@class='css-hfoyj8']", func(e *colly.XMLElement) {
		id := e.ChildText(".//div[@id='description2']//ul/li")
		title := e.ChildText(".//h1[contains(@class,'pdp-product-title')]")
		price := e.ChildText(".//p[contains(@class, 'active-price')]")
		img := e.ChildAttr(".//img[@class='chakra-image']", "src")

		measurements := e.ChildTexts(".//ul[contains(@class, 'measurements-list')]/li")
		materials := e.ChildTexts(".//ul[contains(@class, 'materials-list')]/li")
		features := e.ChildTexts(".//ul[contains(@class, 'features-list')]/li")

		category := e.Request.Ctx.Get("category")

		product := Product{
			ID:           id,
			Title:        title,
			Price:        price,
			Measurements: measurements,
			Materials:    materials,
			Features:     features,
			Img:          img,
			Category:     category,
		}

		*products = append(*products, product)
	})

	err := q.Run(productCollector)
	if err != nil {
		log.Fatal("Queue run failed:", err)
	} else {
		fmt.Println("Queue successfully processed product pages!")
	}
}

func extractCategoryFromURL(url string) string {
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if part == "shop" && i+1 < len(parts) {
			return parts[i+1] // Extract category from URLs like /shop/handbags/view-all
		}
	}
	return "Unknown"
}

//func extractCategoryFromURL(url string) string {
//	parts := strings.Split(url, "/")
//	for i, part := range parts {
//		if part == "shop" && i+1 < len(parts) {
//			return parts[i+1]
//		}
//	}
//	return "Unknown"
//}

// exportToJSON writes the scraped products to a JSON file
func exportToJSON(products []Product, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // Pretty-print JSON with indentation

	if err := encoder.Encode(products); err != nil {
		return fmt.Errorf("could not encode JSON: %v", err)
	}

	return nil
}
