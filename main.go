package main

import (
	"fmt"
	"log"
	"github.com/gocolly/colly/v2"
)

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("www.katespade.com"),
	)

	// Scrape US site
	c.OnHTML(".product-item", func(e *colly.HTMLElement){
		productName := e.ChildText(".product-name")
		productPrice := e.ChildText(".product-price")
		productID := e.ChildAttr("data-product-id", "id")

		fmt.Printf("US Product: %s - Price %s - ID: %s\n", productName, productPrice, productID)
	})


	// Start scraping US site
	err := c.Visit("https://www.katespade.com/us")
	if err != nil {
		log.Fatal(err)
	}


	// Scrape Japan site
	c.OnHTML(".product-item", func(e *colly.HTMLElement){
		productName := e.ChildText(".product-name")
		productPrice := e.ChildText(".product-price")
		productID := e.ChildAttr("data-product-id", "id")

		fmt.Printf("JP Product: %s - Price %s - ID: %s\n", productName, productPrice, productID)
	})

	err = c.Visit("https://www.katespade.com/jp")
	if err != nil {
		log.Fatal(err)
	}

}

