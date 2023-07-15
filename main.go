package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// product declaration
type product struct {
	ItemName           string  `json:"itemName"`
	Retailer           string  `json:"retailer"`
	ProductLink        string  `json:"productLink"`
	ImageLink          string  `json:"imageLink"`
	CurrentPrice       float64 `json:"currentPrice"`
	Rrp                float64 `json:"rrp"`
	DiscountAmount     float64 `json:"discountAmount"`
	DiscountPercentage float64 `json:"discountPercentage"`
}

type refresh struct {
	Refresh string `json:"refresh"`
}

var timeLastUpdated = time.Date(2023, time.January, 1, 0, 0, 0, 0, time.Local)

var realProducts = []product{
	{ItemName: "Toothbrush", Retailer: "Coles", ProductLink: "www.coles.com.au", ImageLink: "www.coles.com.au", CurrentPrice: 7.50, Rrp: 10.00, DiscountAmount: 10.00 - 7.50, DiscountPercentage: 7.50 / 10.00},
	{ItemName: "Mouthwash", Retailer: "Woolworths", ProductLink: "www.woolworths.com.au", ImageLink: "www.woolworths.com.au", CurrentPrice: 8.50, Rrp: 10.00, DiscountAmount: 10.00 - 8.50, DiscountPercentage: 8.50 / 10.00},
	{ItemName: "Chicken Thigh", Retailer: "Woolworths", ProductLink: "www.woolworths.com.au", ImageLink: "www.woolworths.com.au", CurrentPrice: 12.50, Rrp: 13.50, DiscountAmount: 13.50 - 12.50, DiscountPercentage: 12.50 / 13.50},
}
var realProducts2 = []product{
	{ItemName: "Toothbrush", Retailer: "Coles", ProductLink: "www.coles.com.au", ImageLink: "www.coles.com.au", CurrentPrice: 7.50, Rrp: 10.00, DiscountAmount: 10.00 - 7.50, DiscountPercentage: 1 - (7.50 / 10.00)},
	{ItemName: "Mouthwash", Retailer: "Woolworths", ProductLink: "www.woolworths.com.au", ImageLink: "www.woolworths.com.au", CurrentPrice: 8.50, Rrp: 10.00, DiscountAmount: 10.00 - 8.50, DiscountPercentage: 1 - (8.50 / 10.00)},
	{ItemName: "Chicken Thigh", Retailer: "Woolworths", ProductLink: "www.woolworths.com.au", ImageLink: "www.woolworths.com.au", CurrentPrice: 12.50, Rrp: 13.50, DiscountAmount: 13.50 - 12.50, DiscountPercentage: 1 - (12.50 / 13.50)},
}

func main() {
	router := gin.Default()
	router.GET("/products", getProducts)
	router.POST("/refresh", postRefresh)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := router.Run(":" + port); err != nil {
		log.Panicf("error: %s", err)
	}
}

// return `realProducts` as JSON
func getProducts(c *gin.Context) {
	// Check if products updated within 4 hours
	currentTime := time.Now()
	duration := currentTime.Sub(timeLastUpdated)
	if duration >= 4*time.Hour {
		fmt.Printf("Old products! Fetching new data...")
		timeLastUpdated = currentTime
		startWebScrapers()
	}

	sortParam := c.Query("sort")
	sortedProducts := sortProducts(sortParam)

	c.IndentedJSON(http.StatusOK, sortedProducts)
}

func sortProducts(sortMethod string) []product {
	toSort := append([]product{}, realProducts2...)

	switch sortMethod {
	case "lowest-price":
		sort.Slice(toSort, func(i, j int) bool {
			return toSort[i].CurrentPrice < toSort[j].CurrentPrice
		})
	case "highest-percentage":
		sort.Slice(toSort, func(i, j int) bool {
			return toSort[i].DiscountPercentage > toSort[j].DiscountPercentage
		})
	case "biggest-discount-amount":
		sort.Slice(toSort, func(i, j int) bool {
			return toSort[i].DiscountAmount > toSort[j].DiscountAmount
		})
	default:
		sort.Slice(toSort, func(i, j int) bool {
			return toSort[i].DiscountPercentage > toSort[j].DiscountPercentage
		})
	}
	return toSort
}

// if POST receives {refresh: "true"} then fetch new data to hold in memory
func postRefresh(c *gin.Context) {
	var refresh refresh

	if err := c.BindJSON(&refresh); err != nil {
		return
	}

	if refresh.Refresh == "true" {
		startWebScrapers()
		c.IndentedJSON(http.StatusAccepted, "Successfully Refreshed Products")
		return
	} else {
		c.IndentedJSON(http.StatusForbidden, "Forbidden")
		return
	}
}

func startWebScrapers() {
	// TODO: CHANGE THIS TO THE REAL MEMORY ARRAY <-----
	realProducts = nil

	// TODO: retrieve endpoints for web scrapers

	var wg sync.WaitGroup
	urls := []string{"https://jsonplaceholder.typicode.com/posts/1", "https://jsonplaceholder.typicode.com/posts/2", "https://jsonplaceholder.typicode.com/posts/3"}

	for i := range urls {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			fetchProductsFromScraper(urls[i])
		}()

	}
	wg.Wait()
}

func fetchProductsFromScraper(url string) {
	fmt.Printf("Scraping data from: %s\n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	// TODO: append as product type
	// TODO: CHANGE THIS TO THE REAL MEMORY ARRAY <-----
	fmt.Printf("body: %v\n", body)
	// realProducts = append(realProducts, string(body))
}
