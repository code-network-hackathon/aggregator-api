package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

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

// dummy seed data
var products = []product{
	{ItemName: "Toothbrush", Retailer: "Coles", ProductLink: "www.coles.com.au", ImageLink: "www.coles.com.au", CurrentPrice: 7.50, Rrp: 10.00, DiscountAmount: 10.00 - 7.50, DiscountPercentage: 7.50 / 10.00},
	{ItemName: "Mouthwash", Retailer: "Woolworths", ProductLink: "www.woolworths.com.au", ImageLink: "www.woolworths.com.au", CurrentPrice: 8.50, Rrp: 10.00, DiscountAmount: 10.00 - 8.50, DiscountPercentage: 8.50 / 10.00},
	{ItemName: "Chicken Thigh", Retailer: "Woolworths", ProductLink: "www.woolworths.com.au", ImageLink: "www.woolworths.com.au", CurrentPrice: 12.50, Rrp: 13.50, DiscountAmount: 13.50 - 12.50, DiscountPercentage: 12.50 / 13.50},
}

var realProducts = []product{}

func main() {
	router := gin.Default()
	router.GET("/products", getProducts)
	router.POST("/refresh", postRefresh)

	router.Run("localhost:8080")
}

// getProducts responds with all the products as JSON
func getProducts(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, products)
}

// if POST receives {refresh: "true"} then fetch new data to hold in memory
func postRefresh(c *gin.Context) {
	var refresh refresh

	// Call BindJSON to bind the received JSON to refresh
	if err := c.BindJSON(&refresh); err != nil {
		return
	}

	if refresh.Refresh == "true" {
		// TODO: add refresh logic
		startWebScrapers()
		// TODO: save to memory
		c.IndentedJSON(http.StatusAccepted, "Successfully Refreshed Products")
		return
	} else {
		c.IndentedJSON(http.StatusForbidden, "Forbidden")
		return
	}
}

func startWebScrapers() []product {
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

	return products
}

func fetchProductsFromScraper(url string) {
	fmt.Printf("Worker starting %s\n", url)
	// time.Sleep(time.Second)
	// fmt.Printf("Worker done %s \n", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	requestBody := string(body)
	println(requestBody)
	println("^^^^^^^^^^^^^^^")
}
