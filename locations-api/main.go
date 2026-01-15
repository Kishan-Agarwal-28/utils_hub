package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	// "encoding/json"
	// "fmt"
	"log"
	"net/http"
	"os"

	// "strconv"
	"strings"
	"time"

	// "unicode"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	lru "github.com/hashicorp/golang-lru/v2"
	_ "modernc.org/sqlite"
)
var (
	db *sql.DB 
	cache *lru.Cache[string, []Location]
)
func main(){
	var err error
	db, err = sql.Open("sqlite", "./locations.db")
	if err != nil {
		log.Fatal(err)
	}
	cache, err = lru.New[string, []Location](1000)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		log.Println("⚠️ Failed to enable WAL mode:", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	r:= chi.NewRouter()
	r.Use(middleware.Logger)	
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/health"))
	r.Use(middleware.Compress(5))
	r.Use(httprate.Limit(
		100,
		1*time.Minute,
		httprate.WithKeyFuncs(httprate.KeyByIP),
	))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	}))
	
	r.Get("/", documentationHandler)
	r.Get("/api/locations", searchHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}
	log.Printf("🚀 Server running on port %s", port)
	http.ListenAndServe(":"+port, r)
}
func documentationHandler(w http.ResponseWriter, r *http.Request) {
	const apiDocsHTML = `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>City Search API Documentation</title>
		<style>
			:root {
				--bg-color: #f8f9fa;
				--text-color: #212529;
				--accent-color: #007bff;
				--code-bg: #e9ecef;
				--card-bg: #ffffff;
				--border-color: #dee2e6;
			}
			@media (prefers-color-scheme: dark) {
				:root {
					--bg-color: #121212;
					--text-color: #e0e0e0;
					--accent-color: #64b5f6;
					--code-bg: #1e1e1e;
					--card-bg: #1e1e1e;
					--border-color: #333;
				}
			}
			body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif; line-height: 1.6; color: var(--text-color); background: var(--bg-color); margin: 0; padding: 20px; }
			.container { max-width: 800px; margin: 0 auto; background: var(--card-bg); padding: 40px; border-radius: 8px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); border: 1px solid var(--border-color); }
			h1, h2, h3 { color: var(--text-color); }
			h1 { border-bottom: 2px solid var(--accent-color); padding-bottom: 10px; }
			.endpoint { background: var(--code-bg); padding: 10px 15px; border-radius: 5px; font-family: monospace; font-size: 1.1em; display: inline-block; border: 1px solid var(--border-color); }
			.method { color: #fff; background: #28a745; padding: 2px 8px; border-radius: 4px; margin-right: 10px; font-weight: bold; font-size: 0.9em; }
			code { background: var(--code-bg); padding: 2px 5px; border-radius: 4px; font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace; font-size: 0.9em; }
			pre { background: var(--code-bg); padding: 15px; border-radius: 5px; overflow-x: auto; border: 1px solid var(--border-color); }
			table { width: 100%; border-collapse: collapse; margin: 20px 0; }
			th, td { text-align: left; padding: 12px; border-bottom: 1px solid var(--border-color); }
			th { background: var(--code-bg); }
			.badge { display: inline-block; padding: 2px 8px; border-radius: 12px; font-size: 0.8em; font-weight: bold; }
			.required { background: #dc3545; color: white; }
			.optional { background: #6c757d; color: white; }
			.footer { margin-top: 40px; font-size: 0.9em; color: #6c757d; text-align: center; border-top: 1px solid var(--border-color); padding-top: 20px;}
		</style>
	</head>
	<body>

	<div class="container">
		<h1>🌍 City Search API</h1>
		<p>A high-performance, low-latency API to search for global cities, states, and countries. Optimized for speed with in-memory caching and efficient SQL queries.</p>

		<h2>Endpoint</h2>
		<div class="endpoint">
			<span class="method">GET</span> /api/search
		</div>

		<h3>Query Parameters</h3>
		<table>
			<thead>
				<tr>
					<th>Parameter</th>
					<th>Type</th>
					<th>Required</th>
					<th>Description</th>
				</tr>
			</thead>
			<tbody>
				<tr>
					<td><code>city</code></td>
					<td>String</td>
					<td><span class="badge optional">Optional*</span></td>
					<td>Search by city name (e.g. "Paris").</td>
				</tr>
				<tr>
					<td><code>state</code></td>
					<td>String</td>
					<td><span class="badge optional">Optional*</span></td>
					<td>Search by state/province name.</td>
				</tr>
				<tr>
					<td><code>country</code></td>
					<td>String</td>
					<td><span class="badge optional">Optional*</span></td>
					<td>Search by country name.</td>
				</tr>
				<tr>
					<td><code>limit</code></td>
					<td>Integer</td>
					<td><span class="badge optional">No</span></td>
					<td>Max results (Default: <code>10</code>).</td>
				</tr>
			</tbody>
		</table>
		<p><em>* At least one search parameter (city, state, or country) is required.</em></p>

		<h3>Example Request</h3>
		<pre><code>GET /api/search?city=Ashkasham</code></pre>

		<h3>Example Response</h3>
		<pre><code>[
  {
    "city": "Ashkasham",
    "state": "Badakhshan",
    "country": "Afghanistan"
  },
  {
    "city": "Ashkasham 2",
    "state": "Badakhshan",
    "country": "Afghanistan"
  }
]</code></pre>

		<h3>Status Codes</h3>
		<ul>
			<li><code>200 OK</code> - Request successful.</li>
			<li><code>400 Bad Request</code> - Missing query parameters.</li>
			<li><code>429 Too Many Requests</code> - Rate limit exceeded (100 req/min).</li>
			<li><code>500 Internal Server Error</code> - Server-side issue.</li>
		</ul>

		<div class="footer">
			API Version 1.0 &bull; Powered by Go & SQLite &bull; Deployed on Koyeb
		</div>
	</div>

	</body>
	</html>
	`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(apiDocsHTML))
}
type Location struct {
    City    string `json:"city"`
    State   string `json:"state"`
    Country string `json:"country"`
}
func searchHandler(w http.ResponseWriter, r *http.Request) {
    var searchType, searchValue, dbColumn string

    if q := r.URL.Query().Get("city"); q != "" {
        searchType = "city"
        dbColumn = "cities.name"
        searchValue = strings.ToLower(q)
    } else if q := r.URL.Query().Get("state"); q != "" {
        searchType = "state"
        dbColumn = "states.name"
        searchValue = strings.ToLower(q)
    } else if q := r.URL.Query().Get("country"); q != "" {
        searchType = "country"
        dbColumn = "countries.name"
        searchValue = strings.ToLower(q)
    } else {
        http.Error(w, "Missing search parameter (city, state, or country)", http.StatusBadRequest)
        return
    }

    cacheKey := searchType + ":" + searchValue
    
    if locations, found := cache.Get(cacheKey); found {
        respondWithJSON(w, locations, "HIT")
        return
    }

    query := fmt.Sprintf(`
        SELECT cities.name, states.name, countries.name 
        FROM cities
        JOIN states ON cities.state_id = states.id
        JOIN countries ON cities.country_id = countries.id
        WHERE %s LIKE ? 
        ORDER BY length(%s) ASC 
        LIMIT 10`, dbColumn, dbColumn)

    rows, err := db.Query(query, "%"+searchValue+"%")
    if err != nil {
        log.Printf("DB Error: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    results := make([]Location, 0)
    
    for rows.Next() {
        var loc Location
        if err := rows.Scan(&loc.City, &loc.State, &loc.Country); err != nil {
            log.Printf("Scan Error: %v", err)
            continue
        }
        results = append(results, loc)
    }

    cache.Add(cacheKey, results)
    respondWithJSON(w, results, "MISS")
}

func respondWithJSON(w http.ResponseWriter, data any, cacheStatus string) {
    w.Header().Set("X-Cache", cacheStatus)
    w.Header().Set("Cache-Control", "public, max-age=86400")
    w.Header().Set("Content-Type", "application/json")
    
    if err := json.NewEncoder(w).Encode(data); err != nil {
        log.Printf("JSON Encode Error: %v", err)
    }
}