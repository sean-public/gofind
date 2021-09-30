// gofind: Simple in-memory search engine
//
// Scrapes the site(s) you point it at, creates an inverted index
// in memory that can be queried via API. It normalizes text by splitting
// on non-word runes, lowercasing, removing stop words, and then stemming
// with the Snowball algorithm for English.
//
// Configure with ENV vars:
//     PORT (port number, defaults to 8080)
//     START_URL (just one, like: https://your.site/docs)
//     MAX_DEPTH (int: 0 means infinite, 1 means don't follow links at all ...)
//     ALLOWED_DOMAINS (comma separated list of domain names: your.site,foo.co)
//     DISALLOWED_DOMAINS (comma separated list of domains)

package main

import (
    "github.com/gocolly/colly/v2"
    "log"
    "net/http"
    "strings"
    "sync"
    "time"
)

const userAgent = "gofind/1.0"

var docs = make(docCache)
var docCacheLock sync.Mutex

var idx = make(index)
var indexLock sync.Mutex

// addCollectorHandlers accepts a web Collector and attaches callbacks
// for processing HTML as it's parsed async . For ease of chaining, returns
// the modified Collector. Includes capturing errors for logging.
func addCollectorHandlers(c *colly.Collector) *colly.Collector {
    c.OnResponse(func(r *colly.Response) {
        url := r.Request.URL.String()
        log.Println("Processing", url)
        docCacheLock.Lock()
    })
    c.OnHTML("a[href]", func(e *colly.HTMLElement) {
        link := e.Request.AbsoluteURL(e.Attr("href"))
        if _, exists := docs[link]; strings.Contains(link, "http") && !exists {
            e.Request.Visit(link)
        }
    })
    c.OnHTML("title", func(e *colly.HTMLElement) {
        title := e.Text
        url := e.Request.URL.String()
        if doc, ok := docs[url]; ok {
            doc.Title = title
            docs[url] = doc
        } else {
            docs[url] = document{ID: len(docs), Title: title, URL: url}
        }
    })
    c.OnHTML("p,h1,h2,h3,h4,h5,h6,ul,ol,td", func(e *colly.HTMLElement) {
        text := e.Text
        url := e.Request.URL.String()
        if doc, ok := docs[url]; ok {
            doc.Text += " " + text
            docs[url] = doc
        } else {
            docs[url] = document{ID: len(docs), Text: text, URL: url}
        }
    })
    c.OnScraped(func(r *colly.Response) {
        url := r.Request.URL.String()
        if url != "" && len(docs[url].Text) > 2 {
            idx.add(docs[url])
        }
        docCacheLock.Unlock()
    })
    c.OnError(func(r *colly.Response, err error) {
        url := r.Request.URL.String()
        log.Printf("Error processing %v: %v\n", url, err)
    })
    return c
}

func main() {
    initStopwordsList()

    // Initialize the web crawler
    c := addCollectorHandlers(colly.NewCollector())
    c.Async = true
    c.UserAgent = userAgent
    c.MaxDepth = getEnvInt("MAX_DEPTH", 0)
    c.AllowedDomains = getEnvSlice("ALLOWED_DOMAINS")
    c.DisallowedDomains = getEnvSlice("DISALLOWED_DOMAINS")
    c.Limit(&colly.LimitRule{
        DomainGlob:  "*",
        Parallelism: 100,
        Delay:       5 * time.Millisecond,
    })

    // Start crawling, blocking until all goroutines have finished
    startTime := time.Now()
    c.Visit(getEnv("START_URL", ""))
    c.Wait()
    log.Printf("Done crawling %d pages in %s, ready to process queries\n",
        len(docs), time.Since(startTime),
    )

    // Start serving the query endpoint until termination
    http.HandleFunc("/", queryHandler)
    log.Fatalln(http.ListenAndServe(":"+getEnv("PORT", "8080"), nil))
}
