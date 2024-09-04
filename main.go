package main

import (
	//Local Packages
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	//External Packages
	"github.com/PuerkitoBio/goquery"
)

// Type Definitions
type SeoData struct {
	URL             string
	Title           string
	H1              string
	MetaDescription string
	StatusCode      int
}

type Parser interface {
	GetSeoData(resp *http.Response) (SeoData, error)
}

type DefaultParser struct {
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",

	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",

	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",

	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",

	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",

	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}

//Main function

func main() {
	log.Println("Starting the crawler...")
	if err := webCrawlerWrapper(); err != nil {
		log.Fatalf("Error in webCrawlerWrapper: %v", err)
	}
	log.Println("Crawler finished successfully.")
}

func webCrawlerWrapper() error {
	log.Println("Checking internet connection...")
	if err := checkConnection(); err != nil {
		return fmt.Errorf("connection check failed: %v", err)
	}

	urlFlag := flag.String("url", "", "URL to crawl")
	flag.Parse()

	url := *urlFlag
	log.Printf("URL to crawl: %s", url)

	log.Println("Validating URL...")
	if err := validateUrl(url); err != nil {
		return fmt.Errorf("URL validation failed: %v", err)
	}

	// var p Parser = &DefaultParser{}

	log.Println("Starting to scrape sitemaps...")
	results := scrapSiteMaps(url, 10)

	log.Printf("Scraped %d results", len(results))

	file, err := os.Create("Scraped_Site.txt")
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	log.Println("Writing results to file...")
	for _, res := range results {
		_, err := file.WriteString(fmt.Sprintf("URL: %s, Status Code: %d\n", res.URL, res.StatusCode))
		if err != nil {
			log.Printf("Failed to write to file: %v", err)
		}
	}

	log.Println("Finished writing results to file.")
	return nil
}

// Functon Definitions
func checkConnection() error {
	//Make a ping for 5sec to google.com
	_, err := net.DialTimeout("tcp", "google.com:80", 5*time.Second)
	if err != nil {
		return errors.New("no internet connection")
	}
	return nil
}

func validateUrl(url string) error {
	if url == "" {
		return errors.New("URL is empty")
	}
	_, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}
	return nil
}

func extractSiteMapUrls(url string) []string {
	var mu sync.Mutex
	toCrawl := []string{}
	visited := make(map[string]bool)

	var crawl func(url string, depth int)
	crawl = func(url string, depth int) {
		if depth > 10 {
			log.Printf("Max depth reached for URL: %s", url)
			return
		}

		mu.Lock()
		if visited[url] {
			mu.Unlock()
			return
		}
		visited[url] = true
		mu.Unlock()

		log.Printf("Processing URL (depth %d): %s", depth, url)
		response, err := makeRequest(url)
		if err != nil {
			log.Printf("Error retrieving URL: %s, Error: %v", url, err)
			return
		}
		defer response.Body.Close()

		urls, err := extractUrls(response)
		if err != nil {
			log.Printf("Error extracting URLs from response, URL: %s, Error: %v", url, err)
			return
		}

		sitemapFiles, pages := isSiteMap(urls)

		mu.Lock()
		toCrawl = append(toCrawl, pages...)
		mu.Unlock()

		var wg sync.WaitGroup
		for _, sitemapURL := range sitemapFiles {
			wg.Add(1)
			go func(sitemapURL string) {
				defer wg.Done()
				crawl(sitemapURL, depth+1)
			}(sitemapURL)
		}
		wg.Wait()
	}

	timeout := time.After(10 * time.Minute)
	done := make(chan bool)

	go func() {
		crawl(url, 0)
		done <- true
	}()

	select {
	case <-done:
		log.Println("Sitemap extraction completed")
	case <-timeout:
		log.Println("Sitemap extraction timed out after 10 minutes")
	}

	return toCrawl
}

func randomUserAgent() string {

	rand.New(rand.NewSource(time.Now().Unix()))

	randNum := rand.Int() % len(userAgents)

	return userAgents[randNum]
}

func isSiteMap(urls []string) ([]string, []string) {

	sitemapFiles := []string{}

	pages := []string{}

	for _, page := range urls {

		foundSitemap := strings.Contains(page, "xml")

		if foundSitemap {
			fmt.Println("Found Sitemap", page)
			sitemapFiles = append(sitemapFiles, page)
		} else {
			pages = append(pages, page)
		}
	}

	return sitemapFiles, pages
}

var rateLimiter = time.Tick(100 * time.Millisecond)

func makeRequest(url string) (*http.Response, error) {
	<-rateLimiter // Wait for rate limiter

	client := http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", randomUserAgent())

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func extractUrls(response *http.Response) ([]string, error) {

	if response == nil || response.Body == nil {
		return nil, fmt.Errorf("invalid response or response body")
	}
	defer response.Body.Close()

	doc, err := goquery.NewDocumentFromReader(response.Body)

	if err != nil {
		return nil, err
	}

	results := []string{}

	selected := doc.Find("loc")

	for i := range selected.Nodes {
		loc := selected.Eq(i)
		result := loc.Text()
		results = append(results, result)
	}

	return results, nil
}

func scrapeUrls(urls []string, concurrency int) []SeoData {
	var mu sync.Mutex
	var wg sync.WaitGroup

	tokens := make(chan struct{}, concurrency)
	results := []SeoData{}

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			log.Printf("Requesting URL: %s", url)

			res, err := scrapePage(url, tokens)

			if err != nil {
				log.Printf("Error requesting URL: %s, Error: %v", url, err)
			} else {
				mu.Lock()
				results = append(results, res)
				mu.Unlock()
			}
		}(url)
	}
	wg.Wait()
	return results
}

func scrapePage(url string, tokens chan struct{}) (SeoData, error) {
	tokens <- struct{}{}
	defer func() { <-tokens }()

	res, err := makeRequest(url)
	if err != nil {
		return SeoData{}, err
	}
	defer res.Body.Close()

	return SeoData{
		URL:        url,
		StatusCode: res.StatusCode,
	}, nil
}

func crawlPage(url string, tokens chan struct{}) (*http.Response, error) {

	tokens <- struct{}{}

	res, err := makeRequest(url)

	<-tokens

	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return res, err
}

func scrapSiteMaps(url string, concurrency int) []SeoData {
	log.Println("Extracting sitemap URLs...")
	results := extractSiteMapUrls(url)
	log.Printf("Extracted %d URLs from sitemap", len(results))

	if len(results) == 0 {
		log.Println("No URLs extracted from sitemap")
		return []SeoData{}
	}

	log.Println("Scraping URLs...")
	res := scrapeUrls(results, concurrency)
	log.Printf("Scraped %d URLs", len(res))

	return res
}

// Methods

func (p DefaultParser) GetSeoData(res *http.Response) (SeoData, error) {

	doc, err := goquery.NewDocumentFromReader(res.Body)

	if err != nil {
		return SeoData{}, err
	}

	result := SeoData{}

	result.URL = res.Request.URL.String()

	result.StatusCode = res.StatusCode

	result.Title = doc.Find("title").First().Text()

	result.H1 = doc.Find("h1").First().Text()

	result.MetaDescription, _ = doc.Find("meta[name^=description]").Attr("content")

	return result, nil
}
