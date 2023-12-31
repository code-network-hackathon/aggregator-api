package main

import (
	"encoding/json"
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
	ItemName           string `json:"itemName"`
	Retailer           string `json:"retailer"`
	ProductLink        string `json:"productLink"`
	ImageLink          string `json:"imageLink"`
	CurrentPrice       string `json:"currentPrice"`
	Rrp                string `json:"rrp"`
	DiscountAmount     string `json:"discountAmount"`
	DiscountPercentage string `json:"discountPercentage"`
}

type refresh struct {
	Refresh string `json:"refresh"`
}

var timeLastUpdated = time.Date(2023, time.January, 1, 0, 0, 0, 0, time.Local)

var realProducts = []product{}

func main() {
	router := gin.Default()
	router.Use(CORSMiddleware())
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

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
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

	sortedProducts = removeDuplicates(sortedProducts)

	c.IndentedJSON(http.StatusOK, sortedProducts)
}

func sortProducts(sortMethod string) []product {
	// toSort := append([]product{}, realProducts...)
	toSort := realProducts

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
	realProducts = []product{}

	// TODO: retrieve endpoints for web scrapers

	var wg sync.WaitGroup
	urls := []string{"https://aldi-web-scraper.onrender.com/products", "https://coles-web-scraper.onrender.com/products"}

	for i := range urls {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			fetchProductsFromScraper(urls[i])
		}()

		wg.Wait()
	}
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

	// Unmarshal the JSON data into a slice of Product objects
	var products []product
	err = json.Unmarshal(body, &products)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	realProducts = append(realProducts, products...)

}

func removeDuplicates(products []product) []product {
	encountered := map[string]struct{}{} // Map to store encountered names

	// Iterate over the products slice
	result := []product{}
	for _, product := range products {
		if _, ok := encountered[product.ItemName]; !ok {
			// Add the product to the result if the name is not encountered before
			result = append(result, product)
			encountered[product.ItemName] = struct{}{} // Mark the name as encountered
		}
	}

	return result
}
