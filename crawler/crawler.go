package main

import (
	"fmt"
	"log"
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
}

func main() {
	// Initialize the main collector for the homepage
	c := colly.NewCollector(
		colly.AllowedDomains("www.katespade.com"),
	)

	// Initialize the queue for breadth-first crawling
	q, _ := queue.New(
		2, // Number of threads for concurrent crawling
		&queue.InMemoryQueueStorage{MaxSize: 10000},
	)

	// Get all category links from the main page
	categoryLinks := getCategoryLinks(c)

	// Add all found categories to the queue
	for _, category := range categoryLinks {
		fullURL := baseURL + category
		q.AddURL(fullURL)
	}

	// Create a slice to store all products
	var products []Product

	// Initialize the product collector for scraping product details
	productCollector := colly.NewCollector(
		colly.AllowedDomains("www.katespade.com"),
	)

	// Start scraping products
	scrapeProducts(productCollector, q, &products)

	// Output all scraped products
	for _, product := range products {
		fmt.Printf("ID: %s | Title: %s | Price: %s | Measurements: %v | Materials: %v | Features: %v | Img: %s\n",
			product.ID, product.Title, product.Price, product.Measurements, product.Materials, product.Features, product.Img)
	}
}

// getCategoryLinks scrapes the main page to extract category links
func getCategoryLinks(c *colly.Collector) []string {
	var categories []string

	// Scrape the links from the category menu
	c.OnHTML("div.menu-tier-1.css-1h5bbey div.css-wnawyw div.css-19rc50l a", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if strings.HasPrefix(href, "/shop/") && strings.Contains(href, "view-all") {
			categories = append(categories, href)
			fmt.Println("Found category link:", href)
		}
	})

	// Visit the base URL to trigger the OnHTML callbacks
	err := c.Visit(baseURL)
	if err != nil {
		log.Fatal("Failed to visit main page:", err)
	}

	return categories
}

// scrapeProducts handles scraping product listings and individual product details
func scrapeProducts(productCollector *colly.Collector, q *queue.Queue, products *[]Product) {
	// Collect product links from category pages
	productCollector.OnHTML("div.product-tile", func(e *colly.HTMLElement) {
		productLink := e.ChildAttr("div.product-name a", "href")
		if productLink != "" {
			fullProductURL := baseURL + productLink
			fmt.Println("Found product link:", fullProductURL)
			q.AddURL(fullProductURL)
		}
	})

	// Handle pagination to navigate through multiple pages
	productCollector.OnHTML("a.pagination-next", func(e *colly.HTMLElement) {
		nextPage := e.Request.AbsoluteURL(e.Attr("href"))
		fmt.Println("Found next page:", nextPage)
		q.AddURL(nextPage)
	})

	// Scrape details from individual product pages
	productCollector.OnHTML("div.product-detail", func(e *colly.HTMLElement) {
		id := e.ChildAttr("div[data-item-id]", "data-item-id")
		title := e.ChildText("h1.product-name")
		price := e.ChildText("span.product-sales-price")
		img := e.ChildAttr("img.primary-image", "src")

		// Extracting lists of measurements, materials, and features
		measurements := e.ChildTexts("ul.measurements-list li")
		materials := e.ChildTexts("ul.materials-list li")
		features := e.ChildTexts("ul.features-list li")

		product := Product{
			ID:           id,
			Title:        title,
			Price:        price,
			Measurements: measurements,
			Materials:    materials,
			Features:     features,
			Img:          img,
		}

		*products = append(*products, product)
	})

	// Run the queue to process all collected URLs
	err := q.Run(productCollector)
	if err != nil {
		log.Fatal("Queue run failed:", err)
	}
}
