// package main

// import (
// 	"bufio"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	//"math/rand"
// 	"os"
// 	"regexp"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/chromedp/chromedp"
// 	"github.com/joho/godotenv"
// 	"github.com/redis/go-redis/v9"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// // --- Constants ---
// const (
// 	redisExpiry     = 24 * time.Hour
// 	crawlTimeout    = 30 * time.Second
// 	scrollAttempts  = 5
// 	pageLoadDelay   = 2 * time.Second
// )

// // --- Global Variables ---
// var (
// 	db          *gorm.DB
// 	redisClient *redis.Client
// )

// // --- Improved Regex for Product URLs ---
// var productURLPattern = regexp.MustCompile(`(?i)/(dp|gp/product|product|products|item|itm|shop|detail|p)/[a-zA-Z0-9-]+(/|\?|$)`)

// // --- Database Model ---
// type ProductURL struct {
// 	ID     uint   `gorm:"primaryKey"`
// 	Domain string `gorm:"index"`
// 	URL    string `gorm:"unique"`
// }

// // --- Crawl Result Struct ---
// type CrawlResult struct {
// 	Domain string   `json:"domain"`
// 	URLs   []string `json:"urls"`
// }

// // --- Load Domains from File ---
// func loadDomains(filename string) ([]string, error) {
// 	file, err := os.Open(filename)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer file.Close()

// 	var domains []string
// 	scanner := bufio.NewScanner(file)
// 	for scanner.Scan() {
// 		domains = append(domains, strings.TrimSpace(scanner.Text()))
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return nil, err
// 	}
// 	return domains, nil
// }
// //--- Load Environment Variables ---
// func loadEnv() {
// 	err := godotenv.Load(".env")
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}
// }

// func initDB() {
// 	loadEnv() // Load env variables

// 	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
// 		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
// 		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

// 	var err error
// 	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("Database connection failed: %v", err)
// 	}

// 	db.AutoMigrate(&ProductURL{})
// 	log.Println("Database initialized successfully")
// }

// // --- Initialize Redis Client ---
// func initRedis() {
// 	redisAddr := os.Getenv("REDIS_ADDR")
// 	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})

// 	_, err := redisClient.Ping(context.Background()).Result()
// 	if err != nil {
// 		log.Fatalf("Failed to connect to Redis: %v", err)
// 	}

// 	log.Println("Redis connected successfully")
// }

// // --- Extract Product URLs from Page ---
// func extractProductURLs(htmlContent, baseURL string) []string {
// 	matches := productURLPattern.FindAllString(htmlContent, -1)
// 	uniqueURLs := make(map[string]bool)
// 	var productURLs []string

// 	for _, match := range matches {
// 		fullURL := baseURL + match
// 		if !uniqueURLs[fullURL] {
// 			uniqueURLs[fullURL] = true
// 			productURLs = append(productURLs, fullURL)
// 		}
// 	}
// 	return productURLs
// }

// // --- Scrape Product Pages ---
// func scrapeWebsite(url string, resultChan chan<- CrawlResult, wg *sync.WaitGroup) {
// 	defer wg.Done()

// 	opts := chromedp.DefaultExecAllocatorOptions[:]
// 	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
// 	defer cancel()

// 	ctx, cancel := chromedp.NewContext(allocCtx)
// 	defer cancel()

// 	ctx, cancel = context.WithTimeout(ctx, crawlTimeout)
// 	defer cancel()

// 	var htmlContent string
// 	err := chromedp.Run(ctx,
// 		chromedp.Navigate(url),
// 		chromedp.WaitVisible(`body`, chromedp.ByQuery),
// 		chromedp.OuterHTML(`html`, &htmlContent),
// 	)
// 	if err != nil {
// 		log.Printf("Failed to load page: %s | Error: %v", url, err)
// 		return
// 	}

// 	productURLs := extractProductURLs(htmlContent, url)
// 	storeProductURLs(productURLs, url)

// 	resultChan <- CrawlResult{Domain: url, URLs: productURLs}
// }

// // --- Store Product URLs in Database ---
// func storeProductURLs(urls []string, domain string) {
// 	for _, url := range urls {
// 		var count int64
// 		db.Model(&ProductURL{}).Where("url = ?", url).Count(&count)

// 		if count == 0 {
// 			db.Create(&ProductURL{Domain: domain, URL: url})
// 		}
// 	}
// }

// // --- Save Results to JSON File ---
// func saveResults(results []CrawlResult) {
// 	file, err := os.Create("output.json")
// 	if err != nil {
// 		log.Fatalf("Failed to create output file: %v", err)
// 	}
// 	defer file.Close()

// 	jsonData, _ := json.MarshalIndent(results, "", "  ")
// 	file.Write(jsonData)
// 	log.Println("Crawling complete. Results saved in output.json")
// }

// // --- Main Function ---
// func main() {
// 	initDB()
// 	initRedis()

// 	domains, err := loadDomains("domains.txt")
// 	if err != nil {
// 		log.Fatalf("Error loading domains: %v", err)
// 	}

// 	var results []CrawlResult
// 	resultChan := make(chan CrawlResult, len(domains))
// 	var wg sync.WaitGroup

// 	for _, domain := range domains {
// 		wg.Add(1)
// 		go scrapeWebsite(domain, resultChan, &wg)
// 	}

// 	wg.Wait()
// 	close(resultChan)

// 	for res := range resultChan {
// 		results = append(results, res)
// 	}

// 	saveResults(results)
// }


package main

import (
	"bufio"
	"context"
	"encoding/json"
	//"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --- Constants ---
const (
	redisExpiry     = 24 * time.Hour //Cache expiration time for visited URLs
	crawlTimeout    = 90 * time.Second  //Max time for loading a page before timing out
	scrollAttempts  = 5     //Number of times the page scrolls to load more content
	pageLoadDelay   = 4 * time.Second //Wait time after loading pages
)

// --- Global Variables ---
var (
	db          *gorm.DB  //Holds the PostgreSQL database connection
	redisClient *redis.Client // Stores the Redis cache connection
)

// --- Regex Pattern for Product URLs ---
var productURLPattern = regexp.MustCompile(`(?i)/(dp|gp/product|product|products|item|itm|shop|detail|p)/[a-zA-Z0-9-]+(/|\?|$)`)

// --- Database Model ---
//defineing db schema to store unique urls
type ProductURL struct {
	ID     uint   `gorm:"primaryKey"`
	Domain string `gorm:"index"`
	URL    string `gorm:"unique"`
}

// --- Crawl Result Struct ---
//output json format
type CrawlResult struct {
	Domain string   `json:"domain"`
	URLs   []string `json:"urls"`
}

// --- Load Domains from File ---
func loadDomains(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var domains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domains = append(domains, strings.TrimSpace(scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return domains, nil
}

// --- Load Environment Variables (For Local Only) ---
func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
}

// --- Initialize Database Connection ---
func initDB() {
	loadEnv() // Load .env only for local use

	dsn := os.Getenv("DATABASE_URL") // Use Railway's environment variable, here this is only for local use to take .env file
	if dsn == "" {
		log.Fatal("DATABASE_URL not found. Set it in environment variables.")
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Auto-migrate ProductURL table
	db.AutoMigrate(&ProductURL{})
	log.Println("Database connected successfully")
}

// --- Initialize Redis Client ---
func initRedis() {
	loadEnv() // Load .env only for local use

	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		log.Fatal("REDIS_URL not found. Set it in environment variables.")
	}

	redisAddr = strings.TrimPrefix(redisAddr, "redis://") //Trims "redis://" prefix from the Redis connection string

	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
}

// --- Scrape Product Pages ---
func scrapeWebsite(url string, resultChan chan<- CrawlResult, wg *sync.WaitGroup) {
	defer wg.Done() //Marks task completion once this function finishes execution


	//-------------configuring chrome browser for scraping-----------//
	opts := append(chromedp.DefaultExecAllocatorOptions[:],          //initialize default chrome settings
		chromedp.Flag("disable-http2", true),            //to avoids browser compatibility issues
		chromedp.Flag("ignore-certificate-errors", true), // Bypass SSL issues
	)
	//-----------------------------------------------------------//


	//---------------creating browser context ---------//
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...) //Creates a new Chrome browser instance
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx) //to keep browser's current state
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, crawlTimeout) //Prevents infinite waiting for pages that load too slowly
	defer cancel() //to clean resources once scraping done


	//-------------------navigating to web page
	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.Sleep(pageLoadDelay), //Wait for the page to load
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil), // Scroll down to handle infinite scrolling for page that loads dynamically
		chromedp.Sleep(pageLoadDelay), // Wait after scrolling to let content load
		chromedp.OuterHTML(`html`, &htmlContent), //Extracts the updated HTML after scrolling	
	)
	if err != nil {
		log.Printf("Failed to load page: %s | Error: %v", url, err)
		return
	}

	productURLs := extractProductURLs(htmlContent, url)
	storeProductURLs(productURLs, url)
	

	time.Sleep(time.Duration(rand.Intn(3)+2) * time.Second) 	//Randomized delay to prevent blocking

	resultChan <- CrawlResult{Domain: url, URLs: productURLs} //This allows asynchronous processing of scraped data by sending extracted urls to channel
}

// --- Extract Product URLs from Page ---
func extractProductURLs(htmlContent, baseURL string) []string {
	matches := productURLPattern.FindAllString(htmlContent, -1)
	uniqueURLs := make(map[string]bool)
	var productURLs []string

	for _, match := range matches {
		fullURL := baseURL + match
		if !uniqueURLs[fullURL] {
			uniqueURLs[fullURL] = true
			productURLs = append(productURLs, fullURL)
		}
	}
	return productURLs
}

// --- Store Product URLs in Database ---
func storeProductURLs(urls []string, domain string) {
	for _, url := range urls {
		var count int64
		db.Model(&ProductURL{}).Where("url = ?", url).Count(&count)

		if count == 0 {
			db.Create(&ProductURL{Domain: domain, URL: url})
			log.Printf("Stored product URL: %s", url)
		} else {
			log.Printf("Duplicate URL skipped: %s", url)
		}
	}
}

// --- Save Results to JSON File ---
func saveResults(results []CrawlResult) {
	file, err := os.Create("output.json")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	jsonData, _ := json.MarshalIndent(results, "", "  ")
	file.Write(jsonData)
	log.Println("Crawling complete. Results saved in output.json")
}

// --- Main Function ---
func main() {
	// Initialize DB & Redis
	initDB()
	initRedis()

	// Load domains from file
	domains, err := loadDomains("domains.txt")
	if err != nil {
		log.Fatalf("Error loading domains: %v", err)
	}

	var results []CrawlResult
	resultChan := make(chan CrawlResult, len(domains))
	var wg sync.WaitGroup

	for _, domain := range domains {
		wg.Add(1)
		go scrapeWebsite(domain, resultChan, &wg)
	}

	wg.Wait()
	close(resultChan)

	for res := range resultChan {
		results = append(results, res)
	}

	saveResults(results)
}


//https://www.flipkart.com/search?q=shoes
// https://www.flipkart.com/search?q=watch
// https://www.flipkart.com/search?q=mac
// https://www.snapdeal.com/search?keyword=shirt
// https://www.snapdeal.com/search?keyword=shoes
// https://www.snapdeal.com/search?keyword=watch
// https://www.snapdeal.com/search?keyword=mac
// https://www.ebay.com/sch/i.html?_nkw=mac
// https://www.ebay.com/sch/i.html?_nkw=shirt
// https://www.ebay.com/sch/i.html?_nkw=shoes
// https://www.ebay.com/sch/i.html?_nkw=watch