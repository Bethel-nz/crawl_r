# Web Crawler Documentation

## Overview

This program is a web crawler designed to scrape SEO-related data from websites. It starts by crawling a sitemap, extracting URLs, and then scraping individual pages for SEO information.

## Main Components

1. **Main Function**: Entry point of the program.
2. **Web Crawler Wrapper**: Orchestrates the crawling process.
3. **URL Extraction**: Extracts URLs from sitemaps.
4. **Page Scraping**: Scrapes individual pages for SEO data.
5. **Utility Functions**: Various helper functions for making requests, parsing data, etc.

## Detailed Breakdown

### Imports

The program uses both standard library and external packages:

- Standard library: `errors`, `flag`, `fmt`, `log`, `math/rand`, `net`, `net/http`, `os`, `strings`, `sync`, `time`
- External: `github.com/PuerkitoBio/goquery`

### Type Definitions

1. **SeoData**: Struct to store SEO-related information for a page.
2. **Parser**: Interface defining the method for getting SEO data.
3. **DefaultParser**: Struct implementing the Parser interface.

### Global Variables

- `userAgents`: A slice of user agent strings for randomizing HTTP requests.

### Main Function

The `main()` function is the entry point. It calls `webCrawlerWrapper()` and handles any errors.

### Web Crawler Wrapper

`webCrawlerWrapper()` function:

1. Checks internet connection
2. Parses command-line flags
3. Validates the provided URL
4. Scrapes sitemaps
5. Writes results to a file

### URL Extraction

`extractSiteMapUrls()` function:

1. Crawls the sitemap recursively
2. Extracts URLs from the sitemap
3. Handles depth limiting and timeout

### Page Scraping

`scrapeUrls()` function:

1. Concurrently scrapes individual pages
2. Collects SEO data for each page

### Utility Functions

- `checkConnection()`: Checks internet connectivity
- `validateUrl()`: Validates the provided URL
- `randomUserAgent()`: Selects a random user agent
- `isSiteMap()`: Determines if a URL is a sitemap
- `makeRequest()`: Makes an HTTP request with rate limiting
- `extractUrls()`: Extracts URLs from an HTTP response
- `crawlPage()`: Crawls a single page
- `scrapePage()`: Scrapes SEO data from a single page

### Methods

`GetSeoData()`: Method of `DefaultParser` that extracts SEO data from an HTTP response.

## Usage

Run the program with the `-url` flag to specify the website to crawl:

```bash
go run main.go -url https://example.com
```

## Output

The program writes the scraped data to a file named `Scraped_Site.txt` in the current directory.

## Error Handling

The program includes extensive error handling and logging throughout the crawling process.

## Concurrency

The crawler uses Go's concurrency features (goroutines and channels) to perform parallel processing of URLs and implement rate limiting.

## Limitations

- The crawler has a maximum depth limit of 10 for sitemap crawling.
- There's a 10-minute timeout for the entire crawling process.
- The program uses a rate limiter to avoid overwhelming the target server.
