package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	//"math/rand"
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
	redisExpiry     = 24 * time.Hour
	crawlTimeout    = 30 * time.Second
	scrollAttempts  = 5
	pageLoadDelay   = 2 * time.Second
)

// --- Global Variables ---
var (
	db          *gorm.DB
	redisClient *redis.Client
)

// --- Improved Regex for Product URLs ---
var productURLPattern = regexp.MustCompile(`(?i)/(dp|gp/product|product|products|item|itm|shop|detail|p)/[a-zA-Z0-9-]+(/|\?|$)`)

// --- Database Model ---
type ProductURL struct {
	ID     uint   `gorm:"primaryKey"`
	Domain string `gorm:"index"`
	URL    string `gorm:"unique"`
}

// --- Crawl Result Struct ---
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
//--- Load Environment Variables ---
func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func initDB() {
	loadEnv() // Load env variables

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	db.AutoMigrate(&ProductURL{})
	log.Println("Database initialized successfully")
}

// --- Initialize Redis Client ---
func initRedis() {
	redisAddr := os.Getenv("REDIS_ADDR")
	redisClient = redis.NewClient(&redis.Options{Addr: redisAddr})

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
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

// --- Scrape Product Pages ---
func scrapeWebsite(url string, resultChan chan<- CrawlResult, wg *sync.WaitGroup) {
	defer wg.Done()

	opts := chromedp.DefaultExecAllocatorOptions[:]
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, crawlTimeout)
	defer cancel()

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.OuterHTML(`html`, &htmlContent),
	)
	if err != nil {
		log.Printf("Failed to load page: %s | Error: %v", url, err)
		return
	}

	productURLs := extractProductURLs(htmlContent, url)
	storeProductURLs(productURLs, url)

	resultChan <- CrawlResult{Domain: url, URLs: productURLs}
}

// --- Store Product URLs in Database ---
func storeProductURLs(urls []string, domain string) {
	for _, url := range urls {
		var count int64
		db.Model(&ProductURL{}).Where("url = ?", url).Count(&count)

		if count == 0 {
			db.Create(&ProductURL{Domain: domain, URL: url})
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
	initDB()
	initRedis()

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


