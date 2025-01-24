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
	c.OnHTML(".product-tile", func(e *colly.HTMLElement) {
		// fmt.Println(e.DOM.Html())
		productID := e.Attr("data-cnstrc-item-id")
		productTitle := e.Attr("data-cnstrc-item-name")
		productPrice := e.Attr("data-cnstrc-item-price")
	
		fmt.Printf("Product ID: %s - Title: %s - Price: %s\n", productID, productTitle, productPrice)
	})


	// Start scraping US site
	err := c.Visit("https://www.katespade.com/shop/handbags/view-all")
	if err != nil {
		log.Fatal(err)
	}

	c.OnResponse(func(r *colly.Response) {
		fmt.Println(string(r.Body)) // Print the response body
	})

	// Scrape Japan site
	// c.OnHTML(".product-item", func(e *colly.HTMLElement){
	// 	productName := e.ChildText(".product-name")
	// 	productPrice := e.ChildText(".product-price")
	// 	productID := e.ChildAttr("data-product-id", "id")

	// 	fmt.Printf("JP Product: %s - Price %s - ID: %s\n", productName, productPrice, productID)
	// })

	// err = c.Visit("https://www.katespade.com/jp")
	// if err != nil {
	// 	log.Fatal(err)
	// }

}

